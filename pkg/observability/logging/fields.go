package logging

import (
	"go.uber.org/zap"
)

// HTTP fields

func Method(v string) zap.Field {
	return zap.String("http.method", v)
}

func Path(v string) zap.Field {
	return zap.String("http.path", v)
}

func Status(v int) zap.Field {
	return zap.Int("http.status", v)
}

func Duration(v string) zap.Field {
	return zap.String("duration", v)
}

func RequestID(v string) zap.Field {
	return zap.String("request_id", v)
}

func UserID(v string) zap.Field {
	return zap.String("user_id", v)
}
