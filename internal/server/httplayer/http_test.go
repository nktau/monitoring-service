package httplayer

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"github.com/nktau/monitoring-service/internal/server/config"
	"github.com/nktau/monitoring-service/internal/server/storagelayer"
	"github.com/nktau/monitoring-service/internal/server/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var logger = utils.InitLogger()

var cfg config.Config

func TestUpdate(t *testing.T) {
	cfg = config.New()
	storeLayer := storagelayer.New(logger, cfg)
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := New(appLayer, logger, "")
	ts := httptest.NewServer(httpAPI.router)
	defer ts.Close()

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name       string
		targetURL  string
		httpMethod string
		want       want
	}{
		{
			name:       "#1 positive test gauge",
			targetURL:  "/update/gauge/randomMetricName/10.5",
			httpMethod: http.MethodPost,
			want: want{
				code:        200,
				response:    "ok\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "#2 positive test counter",
			targetURL:  "/update/counter/randomMetricName/10",
			httpMethod: http.MethodPost,
			want: want{
				code:        200,
				response:    "ok\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "#3 Invalid metric type",
			targetURL:  "/update/invalidMetricType/randomMetricName/10",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				response:    "wrong metric type\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "#4 Invalid metric value",
			targetURL:  "/update/gauge/randomMetricName/invalidValue",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				response:    "wrong metric value\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "#5 empty metric name",
			targetURL:  "/update/gauge/",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
				response:    "wrong metric name\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "#6 invalid http method",
			targetURL:  "/update/gauge/",
			httpMethod: http.MethodTrace,
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    "",
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(test.httpMethod, ts.URL+test.targetURL, nil)
			if err != nil {
				panic("err")
			}
			res, err := ts.Client().Do(request)
			if err != nil {
				panic("err")
			}
			defer res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, res.StatusCode, test.want.code)
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, string(resBody), test.want.response)
			assert.Equal(t, res.Header.Get("Content-Type"), test.want.contentType)
		})
	}
}

func TestValue(t *testing.T) {
	// create storage layer
	storeLayer := storagelayer.New(logger, cfg)
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := New(appLayer, logger, "")
	ts := httptest.NewServer(httpAPI.router)
	defer ts.Close()
	testMetric := "testMetric"
	testMetricValue := "123.5"
	request, err := http.NewRequest(http.MethodPost, ts.URL+
		fmt.Sprintf("/update/gauge/%s/%s", testMetric, testMetricValue), nil)
	if err != nil {
		panic("Can't post data to server for a test")
	}
	res, err := ts.Client().Do(request)
	if err != nil {
		panic("Can't post data to server for a test")
	}
	defer res.Body.Close()

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name       string
		targetURL  string
		httpMethod string
		want       want
	}{
		{
			name:       "#1 positive test get testMetric value",
			targetURL:  fmt.Sprintf("/value/gauge/%s", testMetric),
			httpMethod: http.MethodGet,
			want: want{
				code:        http.StatusOK,
				response:    testMetricValue + "\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "#2 metric not found",
			targetURL:  "/value/counter/randomMetricName",
			httpMethod: http.MethodGet,
			want: want{
				code:        http.StatusNotFound,
				response:    "metric not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "#3 test for root handler",
			targetURL:  "/",
			httpMethod: http.MethodGet,
			want: want{
				code:        http.StatusOK,
				response:    fmt.Sprintf("<h3>%s: %s</h3>\n", testMetric, testMetricValue),
				contentType: "text/html",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(test.httpMethod, ts.URL+test.targetURL, nil)
			if err != nil {
				panic("err")
			}
			res, err := ts.Client().Do(request)
			if err != nil {
				fmt.Println(err)
				panic("error")
			}
			defer res.Body.Close()
			require.NoError(t, err)
			//res := w.Result()
			// проверяем код ответа
			assert.Equal(t, res.StatusCode, test.want.code)
			// получаем и проверяем тело запроса
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, string(resBody), test.want.response)
			assert.Equal(t, res.Header.Get("Content-Type"), test.want.contentType)
		})
	}
}

func TestHashing(t *testing.T) {
	// create storage layer
	storeLayer := storagelayer.New(logger, cfg)
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := New(appLayer, logger, "ChangeME")
	//ts := httptest.NewServer(httpAPI.router)
	//defer ts.Close()
	requestBody := `{"id":"first","type":"gauge","value":1.1}`
	src := []byte(requestBody)
	h := hmac.New(sha256.New, []byte("ChangeME"))
	h.Write(src)
	expectedHash := h.Sum(nil)
	//fmt.Println(string(dst))
	expectedHashString := fmt.Sprintf("%x", expectedHash)
	fmt.Println("expectedHashString: ", expectedHashString)
	request, err := http.NewRequest(http.MethodPost, "/update/gauge/", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name      string
		body      string
		signature string
		want      want
	}{
		{
			name:      "#1 positive test",
			body:      requestBody,
			signature: expectedHashString,
			want: want{
				code:        http.StatusOK,
				response:    requestBody,
				contentType: "application/json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// create the handler to test, using our custom "next" handler
			handlerToTest := httpAPI.hashing(http.HandlerFunc(httpAPI.updateJSON))

			// create a mock request to use
			//req := httptest.NewRequest("GET", "http://testing", nil)

			// call the handler using a mock response recorder (we'll not use that anyway)

			runRequest := request
			runRequest.Header.Set("HashSHA256", test.signature)
			//fmt.Println("runRequest", runRequest.Header.Get("HashSHA256"))
			//fmt.Println(io.ReadAll(runRequest.Body))
			rec := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rec, runRequest)
			fmt.Println(rec.Code, rec.Body)
			//res, err := ts.Client().Do(runRequest)
			//if err != nil {
			//	fmt.Println(err)
			//	fmt.Println("panic err ahaha")
			//	panic("error")
			//}
			//defer res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, rec.Code, test.want.code)
			resBody, err := io.ReadAll(rec.Body)
			require.NoError(t, err)
			assert.Equal(t, string(resBody), test.want.response)
			//assert.Equal(t, rec. ("Content-Type"), test.want.contentType)
		})
	}
}
