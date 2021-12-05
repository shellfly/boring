package boring

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/shellfly/boring/pkg/log"
)

// Borning copy data between left and right bidirectionally
func Boring(left, right net.Conn) {
	wait := 5 * time.Second
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		if _, err := io.Copy(left, right); err != nil {
			log.Errorln("copy data into tunnel error: ", err)
		}
		right.SetReadDeadline(time.Now().Add(wait))

	}()
	go func() {
		defer wg.Done()
		if _, err := io.Copy(right, left); err != nil {
			log.Errorln("copy data from tunnel error: ", err)
		}
		left.SetReadDeadline(time.Now().Add(wait))
	}()
	wg.Wait()
}
