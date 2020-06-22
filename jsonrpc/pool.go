// TODO: ref https://learnku.com/articles/41137

package jsonrpc

import (
	"container/list"
	"net"
	"sync"
)

type ConnectionPool struct {
	pool list.List
	mu   sync.Mutex
	New  func() (*net.Conn, error)
}

func NewConnPool(newFunc func() (*net.Conn, error)) *ConnectionPool {
	p := &ConnectionPool{New: newFunc}
	p.pool.Init()
	return p
}

func (p *ConnectionPool) Len() int {
	return p.pool.Len()
}

// creat N connection in pool
func (p *ConnectionPool) Init(n int) (int, []error) {
	var errs []error
	var count = 0
	for i := 0; i < n; i++ {
		conn, err := p.New()
		if err != nil {
			errs = append(errs, err)
		}
		p.Put(conn)
		count++
	}
	return count, errs
}

func (p *ConnectionPool) Get() *net.Conn {
	if p.Len() == 0 {
		return nil
	}
	defer p.mu.Unlock()
	p.mu.Lock()
	elm := p.pool.Front()
	if elm == nil {
		return nil
	}
	p.pool.Remove(elm)
	conn, _ := elm.Value.(*net.Conn)
	return conn
}

func (p *ConnectionPool) GetOrNew() (*net.Conn, error) {
	conn := p.Get()
	if conn == nil {
		return p.New()
	}
	return conn, nil
}

func (p *ConnectionPool) Put(conn *net.Conn) {
	defer p.mu.Unlock()
	p.mu.Lock()
	p.pool.PushBack(conn)
}

func ConnClosed(conn *net.Conn) bool {
	//one := make([]byte, 1)
	//conn.SetReadDeadline(time.Now())
	//if _, err := conn.Read(one); err == io.EOF {
	//	l.Printf(logger.LevelDebug, "%s detected closed LAN connection", id)
	//	c.Close()
	//	c = nil
	//} else {
	//	var zero time.Time
	//	c.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	//}
	return false
}
