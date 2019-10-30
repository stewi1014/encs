package encodable_test

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encodable"
)

var stringSink string

var testStrings = []string{
	"John",
	"",
	"a string",
	"abcd",
	"this one has weird characters öääå@ł€®þ®þ↓",
	"a long string 9283gf7n4 5v08297tnd09dn3109h8y1ds32h089nym61d32n096h1d320hn9y861s029386js8179j987+1m2d+87nh9+0917s2j7+91807d4n+9218374dhn098781y2sajm09182y0df9187634+9d718ydsh098736bv0fd217yh54",
	"another long string 23470hnvc507982d3mn+s13248m9+13s2489+7mj1324s+j9m7081324sd+j908m312d+0789jm1d38+2n79h8n97+6ds312986nh+01sd32496hn80+1d345689nh+1d3s24709h+jḿs1473+j029md698h+dn14359h687+0nd1s324h0968n1sd3248jh01s0432jju+980k1342s9+ju08s1342",
	`
	An entire document

Concurrent programming has its own idioms. A good example is timeouts. Although Go's channels do not support them directly, they are easy to implement. Say we want to receive from the channel ch, but want to wait at most one second for the value to arrive. We would start by creating a signalling channel and launching a goroutine that sleeps before sending on the channel:

timeout := make(chan bool, 1)
go func() {
    time.Sleep(1 * time.Second)
    timeout <- true
}()
We can then use a select statement to receive from either ch or timeout. If nothing arrives on ch after one second, the timeout case is selected and the attempt to read from ch is abandoned.

select {
case <-ch:
    // a read from ch has occurred
case <-timeout:
    // the read from ch has timed out
}
The timeout channel is buffered with space for 1 value, allowing the timeout goroutine to send to the channel and then exit. The goroutine doesn't know (or care) whether the value is received. This means the goroutine won't hang around forever if the ch receive happens before the timeout is reached. The timeout channel will eventually be deallocated by the garbage collector.

(In this example we used time.Sleep to demonstrate the mechanics of goroutines and channels. In real programs you should use ' time.After', a function that returns a channel and sends on that channel after the specified duration.)

Let's look at another variation of this pattern. In this example we have a program that reads from multiple replicated databases simultaneously. The program needs only one of the answers, and it should accept the answer that arrives first.

The function Query takes a slice of database connections and a query string. It queries each of the databases in parallel and returns the first response it receives:

func Query(conns []Conn, query string) Result {
    ch := make(chan Result)
    for _, conn := range conns {
        go func(c Conn) {
            select {
            case ch <- c.DoQuery(query):
            default:
            }
        }(conn)
    }
    return <-ch
}
In this example, the closure does a non-blocking send, which it achieves by using the send operation in select statement with a default case. If the send cannot go through immediately the default case will be selected. Making the send non-blocking guarantees that none of the goroutines launched in the loop will hang around. However, if the result arrives before the main function has made it to the receive, the send could fail since no one is ready.

This problem is a textbook example of what is known as a race condition, but the fix is trivial. We just make sure to buffer the channel ch (by adding the buffer length as the second argument to make), guaranteeing that the first send has a place to put the value. This ensures the send will always succeed, and the first value to arrive will be retrieved regardless of the order of execution.

These two examples demonstrate the simplicity with which Go can express complex interactions between goroutines.
`,
}

func TestString(t *testing.T) {
	e := encodable.NewString()
	d := encodable.NewString()
	buff := new(bytes.Buffer)

	for _, str := range testStrings {
		err := e.Encode(unsafe.Pointer(&str), buff)
		if err != nil {
			t.Errorf("encode error: %v", err)
		}

		ns := ""
		err = d.Decode(unsafe.Pointer(&ns), buff)
		if err != nil {
			t.Errorf("decode error: %v", err)
		}

		if ns != str {
			t.Errorf("strings do not match; got %v", ns)
		}
	}
}

func TestBool(t *testing.T) {
	e := encodable.NewBool()
	d := encodable.NewBool()
	buff := new(bytes.Buffer)

	var b, u bool
	b = true
	err := e.Encode(unsafe.Pointer(&b), buff)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}

	err = d.Decode(unsafe.Pointer(&u), buff)
	if err != nil {
		t.Errorf("decode error: %v", err)
	}

	if b != u {
		t.Errorf("encoded %v, got %v", b, u)
	}

	if buff.Len() > 0 {
		t.Errorf("data remaining in buffer")
	}

	b = false
	err = e.Encode(unsafe.Pointer(&b), buff)
	if err != nil {
		t.Errorf("encode error: %v", err)
	}

	err = d.Decode(unsafe.Pointer(&u), buff)
	if err != nil {
		t.Errorf("decode error: %v", err)
	}

	if b != u {
		t.Errorf("encoded %v, got %v", b, u)
	}

	if buff.Len() > 0 {
		t.Errorf("data remaining in buffer")
	}

}

func BenchmarkString(b *testing.B) {
	str := "Hello World!"
	e := encodable.NewString()
	buff := new(bytes.Buffer)

	var j int
	for i := 0; i < b.N; i++ {
		e.Encode(unsafe.Pointer(&str), buff)
		e.Decode(unsafe.Pointer(&stringSink), buff)
		if buff.Len() != 0 {
			b.Fatalf("data still in buffer")
		}

		j++
		if j >= len(testStrings) {
			j = 0
		}
	}
}
