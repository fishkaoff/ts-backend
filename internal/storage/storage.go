package storage

import "errors"

var ErrUserAlreadyExists = errors.New("Этот email уже занят")
var ErrUserNotFound = errors.New("Пользователь с таким email не найден")

var ErrDocumentNotFound = errors.New("Объект не найден")

var ErrBadUserId = errors.New("Невалидный user id")
var ErrBadCartId = errors.New("Невалидный cart id")
var ErrBadProductId = errors.New("Невалдиный product id")
