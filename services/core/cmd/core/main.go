package main

import (
	"context"
	advisorrpc "financialAdvisor/services/core/internal/advisor/delivery/rpc"
	advisorRepo "financialAdvisor/services/core/internal/advisor/infra/pgrepo"
	"financialAdvisor/services/core/internal/advisor/usecase"
	"financialAdvisor/services/core/internal/db"
	financerrpc "financialAdvisor/services/core/internal/finance/delivery/rpc"
	financeRepo "financialAdvisor/services/core/internal/finance/infra/pgrepo"
	advisorv1 "financialAdvisor/services/core/internal/pb/financialadvisor/advisor/v1"
	financev1 "financialAdvisor/services/core/internal/pb/financialadvisor/finance/v1"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func main() {
	pool, err := db.Dbc(context.Background())

	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	//advisor
	userRepo := advisorRepo.NewUserRepo(pool)
	incomeRepo := advisorRepo.NewIncomeRepo(pool)
	debtRepo := advisorRepo.NewDebtRepo(pool)
	advisorHandler := advisorrpc.NewAdvisorHandler(userRepo, incomeRepo, debtRepo, usecase.NewAdvisorUseCase())

	//finance
	movementRepo := financeRepo.NewMovementRepo(pool)
	financeHandler := financerrpc.NewFinanceHandler(movementRepo)

	server := grpc.NewServer()

	//Advisor
	advisorv1.RegisterAdvisorServiceServer(server, advisorHandler)

	//Finance
	financev1.RegisterFinanceServiceServer(server, financeHandler)

	lis, err := net.Listen("tcp", ":"+os.Getenv("PORT"))

	if err != nil {
		log.Fatal(err)
	}

	log.Println("grpc server listener on ", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
