package zapLogger

import "go.uber.org/zap"

func toZapFields(args ...interface{}) (fields []zap.Field) {

	fields = make([]zap.Field, 0, len(args)/2)

	// Convert key-value pairs to zap fields
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue // skip invalid key
		}

		val := args[i+1]

		switch v := val.(type) {
		case string:
			fields = append(fields, zap.String(key, v))
		case int:
			fields = append(fields, zap.Int(key, v))
		case int64:
			fields = append(fields, zap.Int64(key, v))
		case float64:
			fields = append(fields, zap.Float64(key, v))
		case bool:
			fields = append(fields, zap.Bool(key, v))
		default:
			fields = append(fields, zap.Any(key, v))
		}
	}

	return
}
