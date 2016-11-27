package server

import (
	"bufio"
	"log"
	"net"
	"sync"
	"time"
)

type Service struct {
	ch        chan bool
	waitGroup *sync.WaitGroup
}

func NewService() *Service {
	service := &Service{
		ch:        make(chan bool),
		waitGroup: &sync.WaitGroup{},
	}
	service.waitGroup.Add(1)
	return service
}

func (s *Service) Serve(address string) {
	listener, err := newListener(address)
	if err != nil {
		log.Fatalln(err)
	}
	defer s.waitGroup.Done()
	log.Println("Listening on address", address)
	for {
		select {
		case <-s.ch:
			log.Println("Received shutdown message, stopping server on", listener.Addr())
			listener.Close()
			return
		default:
		}
		listener.SetDeadline(time.Now().Add(1e9))
		conn, err := listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println(err)
		}
		log.Println("Client connected from ", conn.RemoteAddr())
		s.waitGroup.Add(1)
		go s.serve(conn)
	}
}

func newListener(address string) (*net.TCPListener, error) {
	tcpAddress, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return &net.TCPListener{}, err
	}
	return net.ListenTCP("tcp", tcpAddress)
}

func (s *Service) Stop() {
	close(s.ch)
	s.waitGroup.Wait()
}

func (s *Service) serve(conn *net.TCPConn) {
	defer conn.Close()
	defer s.waitGroup.Done()
	for {
		select {
		case <-s.ch:
			log.Println("Received shutdown message, disconnecting client from", conn.RemoteAddr())
			return
		default:
		}
		conn.SetDeadline(time.Now().Add(1e9))
		reader := bufio.NewReader(conn)
		buffer, err := reader.ReadBytes('\n')
		if err != nil {
			if opError, ok := err.(*net.OpError); ok && opError.Timeout() {
				continue
			}
			log.Println(err)
		}
		if _, err := conn.Write(buffer); nil != err {
			log.Println(err)
			return
		}
	}
}
