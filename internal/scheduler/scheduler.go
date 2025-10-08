package scheduler

import (
	"database/sql"
	"log"
	"time"

	"github.com/adjaecent/sampatti/internal/repository"
	"github.com/adjaecent/sampatti/internal/service"
)

type Scheduler struct {
	db           *sql.DB
	fetchService *service.FetchService
	planRepo     *repository.PlanRepository
	stopCh       chan struct{}
}

func New(db *sql.DB) *Scheduler {
	planRepo := repository.NewPlanRepository(db)
	userRepo := repository.NewUserRepository(db)
	investmentAccountRepo := repository.NewInvestmentAccountRepository(db)
	fetchRepo := repository.NewFetchRepository(db)
	authService := service.NewAuthService(userRepo, planRepo)
	fetchService := service.NewFetchService(investmentAccountRepo, fetchRepo, authService)

	return &Scheduler{
		db:           db,
		fetchService: fetchService,
		planRepo:     planRepo,
		stopCh:       make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	log.Println("Starting scheduler...")
	
	// Run initial fetch for all plans
	go s.fetchAllPlans()
	
	// Set up 24-hour ticker
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Running scheduled fetch for all plans...")
			go s.fetchAllPlans()
		case <-s.stopCh:
			log.Println("Scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) fetchAllPlans() {
	plans, err := s.getAllPlans()
	if err != nil {
		log.Printf("Failed to get plans for scheduled fetch: %v", err)
		return
	}

	for _, planID := range plans {
		log.Printf("Fetching data for plan %d", planID)
		if err := s.fetchService.FetchAllData(planID); err != nil {
			log.Printf("Failed to fetch data for plan %d: %v", planID, err)
		}
	}
}

func (s *Scheduler) getAllPlans() ([]int, error) {
	return s.planRepo.GetAllPlanIDs()
}