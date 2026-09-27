package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "grpc-homework/proto"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewEntityServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Create
	created, err := c.CreateEntity(ctx, &pb.CreateEntityRequest{
		Name:    "Первая заметка",
		Content: "Привет, gRPC!",
	})
	if err != nil {
		log.Fatalf("CreateEntity: %v", err)
	}
	log.Printf("created: %+v", created.GetEntity())
	id := created.GetEntity().GetId()

	// 2. Get
	got, err := c.GetEntity(ctx, &pb.GetEntityRequest{Id: id})
	if err != nil {
		log.Fatalf("GetEntity: %v", err)
	}
	log.Printf("got: %+v", got.GetEntity())

	// 3. Update
	updated, err := c.UpdateEntity(ctx, &pb.UpdateEntityRequest{
		Id:      id,
		Content: "Обновлённое содержимое",
	})
	if err != nil {
		log.Fatalf("UpdateEntity: %v", err)
	}
	log.Printf("updated: %+v", updated.GetEntity())

	// 4. List
	// добавим ещё пару сущностей
	for i := 0; i < 3; i++ {
		_, _ = c.CreateEntity(ctx, &pb.CreateEntityRequest{
			Name:    "Заметка",
			Content: "текст",
		})
	}
	list, err := c.ListEntities(ctx, &pb.ListEntitiesRequest{Page: 1, PageSize: 10})
	if err != nil {
		log.Fatalf("ListEntities: %v", err)
	}
	log.Printf("total=%d", list.GetTotal())
	for _, e := range list.GetEntities() {
		log.Printf("  - %s | %s | v%d", e.GetId(), e.GetName(), e.GetVersion())
	}

	// 5. Delete
	del, err := c.DeleteEntity(ctx, &pb.DeleteEntityRequest{Id: id})
	if err != nil {
		log.Fatalf("DeleteEntity: %v", err)
	}
	log.Printf("deleted: %v", del.GetSuccess())

	// Проверка, что объекта больше нет
	if _, err := c.GetEntity(ctx, &pb.GetEntityRequest{Id: id}); err != nil {
		log.Printf("expected NotFound: %v", err)
	}
}