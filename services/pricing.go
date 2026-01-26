package services

import (
	"github.com/mongocollectibles/rental-system/models"
)

const (
	MinimumRentalDays = 7
	SpecialRateMultiplier = 2.0
)

// PricingService handles rental fee calculations
type PricingService struct{}

// NewPricingService creates a new pricing service
func NewPricingService() *PricingService {
	return &PricingService{}
}

// CalculateRentalFee calculates the total rental fee based on size and duration
// Returns the daily rate, total fee, and whether special rate was applied
func (s *PricingService) CalculateRentalFee(size models.Size, duration int) (dailyRate float64, totalFee float64, isSpecialRate bool) {
	// Get base daily rate for the size
	baseRate := size.GetDailyRate()
	
	// Determine if special rate applies (duration < minimum)
	isSpecialRate = duration < MinimumRentalDays
	
	// Calculate daily rate
	if isSpecialRate {
		dailyRate = baseRate * SpecialRateMultiplier
	} else {
		dailyRate = baseRate
	}
	
	// Calculate total fee
	totalFee = dailyRate * float64(duration)
	
	return dailyRate, totalFee, isSpecialRate
}

// CalculateQuote generates a rental quote for a collectible
func (s *PricingService) CalculateQuote(collectible *models.Collectible, duration int) models.RentalQuoteResponse {
	dailyRate, totalFee, isSpecialRate := s.CalculateRentalFee(collectible.Size, duration)
	
	return models.RentalQuoteResponse{
		CollectibleID:   collectible.ID,
		CollectibleName: collectible.Name,
		Size:            collectible.Size,
		Duration:        duration,
		DailyRate:       dailyRate,
		TotalFee:        totalFee,
		IsSpecialRate:   isSpecialRate,
	}
}
