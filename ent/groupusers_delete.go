// Code generated by entc, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"

	"github.com/facebookincubator/ent/dialect/sql"
	"github.com/facebookincubator/ent/dialect/sql/sqlgraph"
	"github.com/facebookincubator/ent/schema/field"
	"github.com/huangc28/go-darkpanda-backend/ent/groupusers"
	"github.com/huangc28/go-darkpanda-backend/ent/predicate"
)

// GroupUsersDelete is the builder for deleting a GroupUsers entity.
type GroupUsersDelete struct {
	config
	hooks      []Hook
	mutation   *GroupUsersMutation
	predicates []predicate.GroupUsers
}

// Where adds a new predicate to the delete builder.
func (gud *GroupUsersDelete) Where(ps ...predicate.GroupUsers) *GroupUsersDelete {
	gud.predicates = append(gud.predicates, ps...)
	return gud
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (gud *GroupUsersDelete) Exec(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	if len(gud.hooks) == 0 {
		affected, err = gud.sqlExec(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*GroupUsersMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			gud.mutation = mutation
			affected, err = gud.sqlExec(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(gud.hooks) - 1; i >= 0; i-- {
			mut = gud.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, gud.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// ExecX is like Exec, but panics if an error occurs.
func (gud *GroupUsersDelete) ExecX(ctx context.Context) int {
	n, err := gud.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (gud *GroupUsersDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: groupusers.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: groupusers.FieldID,
			},
		},
	}
	if ps := gud.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return sqlgraph.DeleteNodes(ctx, gud.driver, _spec)
}

// GroupUsersDeleteOne is the builder for deleting a single GroupUsers entity.
type GroupUsersDeleteOne struct {
	gud *GroupUsersDelete
}

// Exec executes the deletion query.
func (gudo *GroupUsersDeleteOne) Exec(ctx context.Context) error {
	n, err := gudo.gud.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{groupusers.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (gudo *GroupUsersDeleteOne) ExecX(ctx context.Context) {
	gudo.gud.ExecX(ctx)
}
