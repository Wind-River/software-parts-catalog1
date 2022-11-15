package generics

import "testing"

func TestBasicQueue(t *testing.T) {
	t.Run("BasicQueue", func(t *testing.T) {
		q := NewQueue[int]()

		q.Push(0)
		q.Push(1)
		q.Push(2)
		q.Push(3)

		if popped := q.Pop(); popped != 0 {
			t.Errorf("Expected 0; got %d", popped)
			t.FailNow()
		}

		q.Push(0)
		q.Push(4)
		q.Push(5)

		if length := q.Length(); length != 6 {
			t.Errorf("Expected Length 6; got %d", length)
			t.FailNow()
		}

		if popped := q.Pop(); popped != 1 {
			t.Errorf("Expected 1; got %d", popped)
			t.FailNow()
		}
		if popped := q.Pop(); popped != 2 {
			t.Errorf("Expected 2; got %d", popped)
			t.FailNow()
		}
		if popped := q.Pop(); popped != 3 {
			t.Errorf("Expected 3; got %d", popped)
			t.FailNow()
		}
		if popped := q.Pop(); popped != 0 {
			t.Errorf("Expected 0; got %d", popped)
			t.FailNow()
		}
		if popped := q.Pop(); popped != 4 {
			t.Errorf("Expected 4; got %d", popped)
			t.FailNow()
		}
		if popped := q.Pop(); popped != 5 {
			t.Errorf("Expected 5; got %d", popped)
			t.FailNow()
		}

		if length := q.Length(); length != 0 {
			t.Errorf("Expected Length 0; got %d", length)
			t.FailNow()
		}
	})
}
