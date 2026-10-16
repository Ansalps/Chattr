package domain

type CreatedPlanDTO struct {
    ID          string
    Name        string
    Amount      int64
    Currency    string
    Period      string
    Interval    float64
    Description string
    Active      bool
}

type CreatedSubscriptionDTO struct {
    ID             string
    PlanID         string
    Status         string
    TotalCount     int
    RemainingCount int
    PaidCount      int
    ShortURL       string // The payment link for the user
}