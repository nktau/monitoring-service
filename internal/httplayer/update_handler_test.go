package httplayer

import (
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
	// а почему без этого не работает?
	// create storage layer
	storeLayer := storagelayer.New()
	// create app layer
	appLayer := applayer.New(storeLayer)
	// create http layer
	httpAPI := New(appLayer)

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name      string
		targetUrl string
		want      want
	}{
		{
			name:      "#1 positive test gauge",
			targetUrl: "/update/gauge/randomMetricName/10.5",
			want: want{
				code:        200,
				response:    "ok\n",
				contentType: "text/plain",
			},
		},
		{
			name:      "#2 positive test counter",
			targetUrl: "/update/counter/randomMetricName/10",
			want: want{
				code:        200,
				response:    "ok\n",
				contentType: "text/plain",
			},
		},
		{
			name:      "#3 Invalid metric type",
			targetUrl: "/update/invalidMetricType/randomMetricName/10",
			want: want{
				code:        http.StatusBadRequest,
				response:    "wrong metric type\n",
				contentType: "text/plain; charset=utf-8", // Это норм, что пришлось добавить charset=utf-8?
			},
		},
		{
			name:      "#4 Invalid metric value",
			targetUrl: "/update/gauge/randomMetricName/invalidValue",
			want: want{
				code:        http.StatusBadRequest,
				response:    "wrong metric value\n",
				contentType: "text/plain; charset=utf-8", // Это норм, что пришлось добавить charset=utf-8?
			},
		},
		{
			name:      "#5 empty metric name",
			targetUrl: "/update/gauge/",
			want: want{
				code:        http.StatusNotFound,
				response:    "wrong metric name\n",
				contentType: "text/plain; charset=utf-8", // Это норм, что пришлось добавить charset=utf-8?
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.targetUrl, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			httpAPI.update(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, res.StatusCode, test.want.code)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, string(resBody), test.want.response)
			assert.Equal(t, res.Header.Get("Content-Type"), test.want.contentType)
		})
	}
}
