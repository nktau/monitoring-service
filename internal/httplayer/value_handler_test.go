package httplayer

import (
	"fmt"
	"github.com/nktau/monitoring-service/internal/applayer"
	"github.com/nktau/monitoring-service/internal/storagelayer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValue(t *testing.T) {

	// create storage layer
	storeLayer := storagelayer.New()
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := New(appLayer)
	ts := httptest.NewServer(httpAPI.router)
	defer ts.Close()
	testMetric := "testMetric"
	testMetricValue := "123.500000"
	request, err := http.NewRequest(http.MethodPost, ts.URL+
		fmt.Sprintf("/update/gauge/%s/%s", testMetric, testMetricValue), nil)
	if err != nil {
		panic("Can't post data to server for a test")
	}
	_, err = ts.Client().Do(request)
	if err != nil {
		panic("Can't post data to server for a test")
	}

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
				contentType: "text/plain",
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
