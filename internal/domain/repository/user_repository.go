package repository

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
	Create(ctx context.Context, email string) (int, error)
}
