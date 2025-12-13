package di

func DependencyIndjection(cfg *config.Config) (*services.AuthSubscriptionServer, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}
	PostRelationRepository := repository.NewPostRelationRepository(gormDB)
	SmtpUtil := smtp.NewSmtpUtil(&cfg.Smtp)
	JwtUtil := jwt.NewJwtUtil()
	RandomUtil := randomnumber.NewRandomNumberUtil()
	AwsS3Client, err := AwsS3.NewS3Client(cfg.Aws.AwsAccessKey, cfg.Aws.AwsSecretAccessKey, cfg.Aws.AwsRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize s3 client: %w", err)
	}
	razorpayClient := utils.NewRazorpayClient(cfg.Razorpay.KeyId, cfg.Razorpay.KeySecret)
	AuthSubscriptionUsecase := usecase.NewAuthSubscriptionUsecase(AuthSubscriptionRepository,
		RandomUtil, SmtpUtil, &cfg.Token, JwtUtil, razorpayClient, AwsS3Client,cfg.Aws.AwsBucket)
	AuthSubscriptionServiceServer := services.NewAuthSubscriptionServer(AuthSubscriptionUsecase)

	return AuthSubscriptionServiceServer, nil
}