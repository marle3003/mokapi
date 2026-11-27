package git

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(r *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTransport_RoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	t.Run("request to another repo goes to the default transport", func(t *testing.T) {
		tr := newTransport()
		tr.Add("https://github.com/org/auth", roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			require.Fail(t, "must not be used")
			return nil, nil
		}))

		req, err := http.NewRequest(http.MethodGet, srv.URL+"/other", nil)
		require.NoError(t, err)
		res, err := tr.RoundTrip(req)
		require.NoError(t, err)
		defer res.Body.Close()
		require.Equal(t, http.StatusAccepted, res.StatusCode)
	})

	t.Run("matching prefix uses its transport", func(t *testing.T) {
		tr := newTransport()
		tr.Add(srv.URL+"/auth", roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusTeapot, Body: http.NoBody}, nil
		}))

		req, err := http.NewRequest(http.MethodGet, srv.URL+"/auth/repo", nil)
		require.NoError(t, err)
		res, err := tr.RoundTrip(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusTeapot, res.StatusCode)
	})

	t.Run("without transports goes to the default transport", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
		require.NoError(t, err)
		res, err := newTransport().RoundTrip(req)
		require.NoError(t, err)
		defer res.Body.Close()
		require.Equal(t, http.StatusAccepted, res.StatusCode)
	})

	t.Run("parallel requests while adding transports", func(t *testing.T) {
		tr := newTransport()
		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				tr.Add("https://example.org/"+string(rune('a'+i)), roundTripperFunc(nil))
			}()
			go func() {
				defer wg.Done()
				req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
				res, err := tr.RoundTrip(req)
				if err == nil {
					res.Body.Close()
				}
				require.NoError(t, err)
			}()
		}
		wg.Wait()
	})
}
