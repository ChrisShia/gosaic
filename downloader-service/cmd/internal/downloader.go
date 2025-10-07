package internal

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/ChrisShia/jsonlog"
	"github.com/nats-io/nats.go"
)

const chunkSize = 20

type Downloader struct {
	By     http.Client
	To     To
	logger *jsonlog.Logger
	quit   chan struct{}
}

func NewDownloader(save func(string, io.Reader), logger *jsonlog.Logger) *Downloader {
	return &Downloader{
		By:     http.Client{},
		To:     save,
		logger: logger,
	}
}

type To func(key string, input io.Reader)

func (d *Downloader) listenForOsSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	defer signal.Stop(c)

	sig := <-c
	switch sig {
	case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
		d.quit <- struct{}{}
	}
}

// DownloadN
// TODO:
// Add error return to mitigate errors, since errors may occur but the http response
// does not show anything.
func (d *Downloader) DownloadN(natsClient *nats.Conn, requestorIp string, n int, request *http.Request) {
	wg := &sync.WaitGroup{}
	wg.Add(n)

	_, err := natsClient.Subscribe("downloads", func(msg *nats.Msg) {
		go func() {
			defer wg.Done()
			d.Store(requestorIp, msg.Data)
		}()
	})
	if err != nil {
		d.logger.PrintError(err, nil)
	}

	chunks, mod := chunksAndRemainder(n)

	if chunks > 0 {
		for i := chunks; i > 0; i-- {
			go func() {
				for j := 1; j <= chunkSize; j++ {
					d.Get(natsClient, request)
				}
			}()
		}
	}
	if n > chunkSize && mod > 0 {
		go func() {
			for j := 1; j <= mod; j++ {
				d.Get(natsClient, request)
			}
		}()
	}

	wg.Wait()
}

func chunksAndRemainder(n int) (int, int) {
	mod := n % chunkSize
	chunks := n / chunkSize
	return chunks, mod
}

func (d *Downloader) Get(nc *nats.Conn, req *http.Request) {
	response, err := d.By.Do(req)
	if err != nil {
		d.logger.PrintError(err, map[string]string{
			"request": req.URL.String(),
		})
	}
	defer response.Body.Close()

	bs, err := io.ReadAll(response.Body)
	if err != nil {
		d.logger.PrintError(err, map[string]string{
			"response":             response.Status,
			"response_body":        string(bs),
			"response_body_length": strconv.Itoa(len(bs)),
		})
	}

	err = nc.Publish("downloads", bs)
	if err != nil {
		d.logger.PrintError(err, nil)
	}
}

func (d *Downloader) Store(key string, bs []byte) {
	dataReader := bytes.NewReader(bs)

	if dataReader.Len() == 0 {
		d.logger.PrintWarning("Empty file, ignored.", nil)
		return
	}

	d.To(key, dataReader)
}
