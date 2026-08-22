package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	// The client method is the Go spelling of infrai.flags.rollout.
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		panic("INFRAI_API_KEY is required")
	}
	client := NewClient(key)
	ctx := context.Background()
	if err := client.SetFlag(ctx, "health-records-v2", false, true); err != nil {
		panic(err)
	}
	if err := client.Rollout(ctx, "health-records-v2", 10); err != nil {
		panic(err)
	}
	value, err := client.GetValue(ctx, "health-records-v2")
	if err != nil {
		panic(err)
	}
	fmt.Printf("health-records-v2 rollout configured; default_value=%t\n", value.DefaultValue)
}
