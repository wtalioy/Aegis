package analysis_test

import (
	"context"
	"testing"
	"time"

	"aegis/internal/analysis/types"
	aiservice "aegis/internal/platform/ai/service"
	"aegis/internal/platform/storage"
	"aegis/internal/system"
	"aegis/internal/telemetry/workload"
	"aegis/tests/fakes"
)

func TestAIClient_DiagnoseChatAndStreamUseFakeProvider(t *testing.T) {
	provider := fakes.NewAIProvider()
	provider.SingleResponse = "diagnosis"
	provider.MultiResponse = "chat response"
	provider.StreamResponses = []string{"hello ", "world"}

	service := aiservice.NewClient(provider)
	stats := system.NewStats(10, time.Second)
	workloads := workload.NewRegistry(10)
	store := storage.NewManager(10, 10)

	diagnosis, err := service.Diagnose(context.Background(), stats, workloads, store, nil)
	if err != nil {
		t.Fatalf("diagnose: %v", err)
	}
	if diagnosis.Analysis != "diagnosis" {
		t.Fatalf("unexpected diagnosis response: %+v", diagnosis)
	}

	chat, err := service.Chat(context.Background(), "session-1", "what happened?", stats, workloads, store, nil)
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if chat.Message != "chat response" {
		t.Fatalf("unexpected chat response: %+v", chat)
	}

	stream, err := service.ChatStream(context.Background(), "session-1", "stream?", stats, workloads, store, nil)
	if err != nil {
		t.Fatalf("chat stream: %v", err)
	}
	var content string
	for token := range stream {
		content += token.Content
	}
	if content != "hello world" {
		t.Fatalf("unexpected stream content: %q", content)
	}

	history := service.GetChatHistory("session-1")
	if len(history) == 0 {
		t.Fatal("expected chat history to be recorded")
	}
	if history[len(history)-1].Role != "assistant" {
		t.Fatalf("unexpected chat history tail: %+v", history[len(history)-1])
	}

	status := service.GetStatus()
	if status.Provider != "fake-ai" || !status.IsLocal {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestAIClient_AskAboutInsightUsesSingleChatProvider(t *testing.T) {
	provider := fakes.NewAIProvider()
	provider.SingleResponse = "insight answer"
	service := aiservice.NewClient(provider)

	response, err := service.AskAboutInsight(context.Background(), &types.AskInsightRequest{})
	if err != nil {
		t.Fatalf("ask about insight: %v", err)
	}
	if response.Answer != "insight answer" {
		t.Fatalf("unexpected insight response: %+v", response)
	}
}
