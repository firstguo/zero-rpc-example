package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "zero-rpc-example/buf_proto_example/gen/go/tripo/user/v1"
	"zero-rpc-example/internal/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	target := fmt.Sprintf("%s://127.0.0.1:2379/pro/tripo.user.v1.rpc", common.SchemeDiscovJSON)

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, common.Name)),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	// Test 1: Request with matching tag
	fmt.Println("=== Test 1: Request with matching tag (pfb=user-enhanced) ===")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()
	ctx1 = common.WithTags(ctx1, map[string]string{"pfb": "user-enhanced"})
	resp1, err := client.GetUser(ctx1, &pb.GetUserRequest{UserId: "123"})
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		fmt.Printf("Response: %+v\n", resp1)
	}

	// Test 2: Request with non-matching tag (should timeout)
	fmt.Println("\n=== Test 2: Request with non-matching tag (pfb=other-db) ===")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	ctx2 = common.WithTags(ctx2, map[string]string{"pfb": "other-db"})
	resp2, err := client.GetUser(ctx2, &pb.GetUserRequest{UserId: "456"})
	if err != nil {
		log.Printf("Request failed (expected - no matching server): %v", err)
	} else {
		fmt.Printf("Response: %+v\n", resp2)
	}

	// Test 3: Request without tags (should match all)
	fmt.Println("\n=== Test 3: Request without tags ===")
	ctx3, cancel3 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel3()
	resp3, err := client.GetUser(ctx3, &pb.GetUserRequest{UserId: "789"})
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		fmt.Printf("Response: %+v\n", resp3)
	}
}
