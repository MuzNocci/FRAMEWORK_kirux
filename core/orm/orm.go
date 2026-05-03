// Package orm fornece um ORM leve e fluente para o Kyrux Framework.
//
// Design: generics + reflection cacheada + SQL explícito com placeholders.
// Suporta multi-database e multi-tenant via database.DB.Schema.
//
// Uso básico:
//
//	type User struct {
//	    ID   int64  `kyrux:"pk"`
//	    Name string `kyrux:"size:100"`
//	}
//
//	// SELECT
//	users, err := orm.From[User](db).Where("active = ?", true).OrderBy("id DESC").Limit(20).All()
//
//	// INSERT
//	user := User{Name: "Maria"}
//	err = orm.Create(db, &user)
//
//	// UPDATE
//	err = orm.From[User](db).Where("id = ?", 1).Update(map[string]any{"name": "Carlos"})
//
//	// DELETE
//	err = orm.From[User](db).Where("id = ?", 1).Delete()
//
//	// Multi-tenant (schema por request)
//	db := fw.DB.Use().WithSchema("tenant_abc")
//	users, err := orm.From[User](db).All()
package orm

import "kyrux/core/database"

// From inicia um Query builder fluente para o tipo T.
// T deve ser um struct concreto com campos exportados.
func From[T any](db *database.DB) *Query[T] {
	return &Query[T]{
		db:   db,
		meta: metaOf[T](),
	}
}
