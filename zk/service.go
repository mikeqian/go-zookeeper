package zk

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

var (
	ErrRegisterTwice = errors.New("zk: trying to register twice")
)

type Service struct {
	c           *Conn
	path        string
	acl         []ACL
	servicePath string
	seq         int
}

// NewLock creates a new lock instance using the provided connection, path, and acl.
// The path must be a node that is only used by this lock. A lock instances starts
// unlocked until Lock() is called.
func NewService(c *Conn, path string, acl []ACL) *Service {
	return &Service{
		c:    c,
		path: path,
		acl:  acl,
	}
}

func (s *Service) Register(url string) error {
	if s.servicePath != "" {
		return ErrRegisterTwice
	}

	prefix := fmt.Sprintf("%s/service-", s.path)

	path := ""
	var err error
	for i := 0; i < 3; i++ {
		path, err = s.c.CreateProtectedEphemeralSequential(prefix, []byte(url), s.acl)
		if err == ErrNoNode {
			// Create parent node.
			parts := strings.Split(s.path, "/")
			pth := ""
			for _, p := range parts[1:] {
				pth += "/" + p
				_, err := s.c.Create(pth, []byte{}, 0, s.acl)
				if err != nil && err != ErrNodeExists {
					return err
				}
			}
		} else if err == nil {
			break
		} else {
			return err
		}
	}

	if err != nil {
		return err
	}

	seq, err := parseSeq(path)
	if err != nil {
		return err
	}

	for {
		log.Println(s.path)
		children, _, err := s.c.Children(s.path)
		if err != nil {
			return err
		}

		lowestSeq := seq
		prevSeq := 0
		prevSeqPath := ""
		for _, p := range children {
			s, err := parseSeq(p)
			if err != nil {
				return err
			}
			if s < lowestSeq {
				lowestSeq = s
			}
			if s < seq && s > prevSeq {
				prevSeq = s
				prevSeqPath = p
			}
		}

		if seq == lowestSeq {
			// Acquired the lock
			break
		}

		// Wait on the node next in line for the lock
		_, _, ch, err := s.c.GetW(s.path + "/" + prevSeqPath)
		if err != nil && err != ErrNoNode {
			return err
		} else if err != nil && err == ErrNoNode {
			// try again
			continue
		}

		ev := <-ch
		if ev.Err != nil {
			return ev.Err
		}
	}

	s.seq = seq
	s.servicePath = path
	return nil
}
