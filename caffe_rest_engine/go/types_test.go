package main

import (
	"fmt"
	"testing"
)

const (
	MODEL_NAME = "model"
	IMAGE_NAME = "image"
	NUM_MODELS = 10
)

func makeName(m string, num int) string {
	return fmt.Sprintf("%s-%d", m, num)
}

func TestSimpleHLL(t *testing.T) {
	hll := NewHashyLinkedList()

	for i := 0; i < NUM_MODELS; i++ {
		for j := 0; j < 3*MAX_BATCH_AMT; j++ {
			j := Job{
				makeName(MODEL_NAME, i),
				makeName(IMAGE_NAME, j),
				nil,
			}
			hll.AddJob(j)
		}
	}

	for i := 0; i < NUM_MODELS; i++ {
		for j := 0; j < 3; j++ {
			model := makeName(MODEL_NAME, i)
			result := hll.PopFront(MAX_BATCH_AMT)
			if result == nil {
				t.Errorf("Result of PopFront is nil: %s", model)
			} else {
				if len(result) != MAX_BATCH_AMT {
					t.Errorf("Empty result: %s", model)
				} else {
					for k := 0; k < MAX_BATCH_AMT; k++ {
						if result[k].Image != makeName(IMAGE_NAME, k+(j*MAX_BATCH_AMT)) {
							t.Errorf(
								"Comp failed: %s %s %s",
								result[k].Model, result[k].Image,
								makeName(IMAGE_NAME, k+(j*MAX_BATCH_AMT)),
							)
						}

					}
				}

			}
		}
	}
}
