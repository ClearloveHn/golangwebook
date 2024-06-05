package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ClearloveHn/golangwebook/webook/internal/repository"
	"github.com/ClearloveHn/golangwebook/webook/internal/service/sms"
	"math/rand"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany

//go:generate mockgen -source=./code.go -package=svcmocks -destination=./mocks/code.mock.go CodeService
type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context,
		biz, phone, inputCode string) (bool, error)
}

type codeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo: repo,
		sms:  smsSvc,
	}
}

// Send 方法用于发送验证码
func (svc *codeService) Send(ctx context.Context, biz, phone string) error {

	// 验证码进缓存
	code := svc.generate()
	err := svc.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}

	const codeTplId = "1877556"
	return svc.sms.Send(ctx, codeTplId, []string{code}, phone)
}

// Verify 方法用于验证验证码
func (svc *codeService) Verify(ctx context.Context,
	biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if errors.Is(err, repository.ErrCodeVerifyTooMany) {
		// 相当于，我们对外面屏蔽了验证次数过多的错误，我们就是告诉调用者，你这个不对
		return false, nil
	}
	return ok, err
}

func (svc *codeService) generate() string {
	// 0-999999
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
