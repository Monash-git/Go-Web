package service

import (
	"backend/internal/repository"
	"backend/internal/service/sms"
	"context"
	"fmt"
	"math/rand"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany

type CodeService interface{
	Send(ctx context.Context, biz, phone string)error
	Verify(ctx context.Context, biz, phone, inputcode string) (bool,error)
}

type codeService struct{
	repo repository.CodeRepository
	sms sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo: repo,
		sms: smsSvc,
	}
}

func (svc *codeService) generate() string {
	//0-999999
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d",code)
}

func (svc *codeService) Send(ctx context.Context, biz, phone string)error{
	code := svc.generate()
	err := svc.repo.Set(ctx,biz,phone,code)
	if err != nil {
		return err
	}
	const codeTplId = "1877556"
	return svc.sms.Send(ctx,codeTplId,[]string{code},phone)
}

func (svc *codeService) Verify(ctx context.Context, biz, phone, inputcode string) (bool,error) {
	ok, err := svc.repo.Verify(ctx,biz,phone,inputcode)
	if err == repository.ErrCodeVerifyTooMany{
		//相当于频闭了验证次数过多的错误，只告诉调用者，你的验证码错误
		return false,nil
	}
	return ok, err
}