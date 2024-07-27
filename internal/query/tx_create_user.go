package query

import "context"

type CreateUserTxParams struct {
	AfterCreate func(user *User) error
	CreateUserParams
}

type CreateUserTxResult struct {
	User *User
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (*CreateUserTxResult, error) {
	result := &CreateUserTxResult{}

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.User)
	})

	return result, err
}
