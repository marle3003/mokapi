package buffer_test

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"mokapi/buffer"
	"strings"
	"testing"
)

func TestPageBuffer(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Write",
			test: func(t *testing.T) {
				pb := buffer.NewPageBuffer()
				n, err := pb.Write([]byte("hello world"))
				require.NoError(t, err)
				require.Equal(t, 11, n)
				require.Equal(t, 11, pb.Size())
			},
		},
		{
			name: "WriteAt",
			test: func(t *testing.T) {
				pb := buffer.NewPageBuffer()
				_, _ = pb.Write([]byte("hello world"))
				pb.WriteAt([]byte("W"), 6)
				require.Equal(t, 11, pb.Size())
				b := new(bytes.Buffer)
				n, err := pb.WriteTo(b)
				require.NoError(t, err)
				require.Equal(t, 11, n)
				require.Equal(t, "hello World", b.String())
			},
		},
		{
			name: "should use pool",
			test: func(t *testing.T) {
				// init
				pb := buffer.NewPageBuffer()
				data := []byte("hello world")
				_, _ = pb.Write(data)

				n := testing.AllocsPerRun(10, func() {
					pb.Unref()
					pb = buffer.NewPageBuffer()
					_, _ = pb.Write(data)
				})

				//
				require.Equal(t, float64(0), n, "expected 0 allocations per run")
			},
		},
		{
			name: "test page size",
			test: func(t *testing.T) {
				data := make([]byte, 65537)
				pb := buffer.NewPageBuffer()
				_, err := pb.Write(data)
				require.NoError(t, err)
				require.Equal(t, 65537, pb.Size())
			},
		},
		{
			name: "using pages with slice",
			test: func(t *testing.T) {
				data := make([]byte, 100000)
				pb := buffer.NewPageBuffer()
				_, err := pb.Write(data)
				require.NoError(t, err)
				f := pb.Slice(65537, 70000)
				require.Equal(t, 4463, f.Size())

				n := testing.AllocsPerRun(1, func() {
					pb.Unref()
					pb = buffer.NewPageBuffer()
					_, _ = pb.Write(data)
				})
				// expected 2, one for the page used by fragment and one for the buffer array in the page
				require.Equal(t, float64(2), n, "expected 2 allocations")
			},
		},
		{
			name: "read from fragment",
			test: func(t *testing.T) {
				data := []byte(strings.Repeat("a", 100000))
				pb := buffer.NewPageBuffer()
				_, err := pb.Write(data)
				require.NoError(t, err)
				f := pb.Slice(65537, 70000)
				require.Equal(t, 4463, f.Size())
				b := make([]byte, 4463)
				n, err := f.Read(b)
				require.NoError(t, err)
				require.Equal(t, 4463, n)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
