package domain

// Эта структура решает классическую проблему «трёх состояний» (tri-state) — чаще всего в API при обработке
// PATCH-запросов:
// Поле не передано вовсе (Set: false).
// Поле явно передано как null (Set: true, Value: nil).
// Поле передано со значением (Set: true, Value: &val).
type Nullable[T any] struct {
	Value *T
	Set   bool
}
