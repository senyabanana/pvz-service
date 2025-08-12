package entity

import "errors"

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrEmailTaken             = errors.New("email already taken")
	ErrInvalidCity            = errors.New("invalid city")
	ErrPVZNotFound            = errors.New("pvz not found")
	ErrReceptionAlreadyExists = errors.New("open reception already exists")
	ErrNoActiveReception      = errors.New("no active reception for this PVZ")
	ErrInvalidProductType     = errors.New("invalid product type")
	ErrInvalidUserRole        = errors.New("invalid user role")
	ErrNoOpenReception        = errors.New("no open receptions")
	ErrNoProductsToDelete     = errors.New("no product to delete")
	ErrReceptionAlreadyClosed = errors.New("reception already closed")
)
