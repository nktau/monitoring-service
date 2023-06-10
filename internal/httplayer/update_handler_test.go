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

func TestUpdate(t *testing.T) {

	// create storage layer
	storeLayer := storagelayer.New()
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := New(appLayer)
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
				contentType: "text/plain",
			},
		},
		{
			name:       "#2 positive test counter",
			targetURL:  "/update/counter/randomMetricName/10",
			httpMethod: http.MethodPost,
			want: want{
				code:        200,
				response:    "ok\n",
				contentType: "text/plain",
			},
		},
		{
			name:       "#3 Invalid metric type",
			targetURL:  "/update/invalidMetricType/randomMetricName/10",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				response:    "wrong metric type\n",
				contentType: "text/plain; charset=utf-8", // Это норм, что пришлось добавить charset=utf-8?
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
				contentType: "text/plain; charset=utf-8", // Это норм, что пришлось добавить charset=utf-8?
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
			// создаём новый Recorder
			//w := httptest.NewRecorder()
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
