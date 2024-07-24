package hyperliquid

import (
	"context"
	"fmt"
)

type Logger interface {
	LogInfo(ctx context.Context, msg string)
	LogErr(ctx context.Context, msg string, err error)
}

type DefaultLogger struct {
}

func (d *DefaultLogger) LogInfo(ctx context.Context, msg string) {
	fmt.Printf("%s\n", msg)
}

func (d *DefaultLogger) LogErr(ctx context.Context, msg string, err error) {
	fmt.Printf("%s\n. Err %s", msg, err.Error())
}
