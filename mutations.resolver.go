
const (
	contentType = "application/octet-stream"
	validFor    = time.Hour // URL valid for 1 hour
)

// StorageLayer defines a struct with a GCS client.
type StorageLayer struct {
	client *storage.Client
}

// NewStorageLayer initializes a new StorageLayer with a GCS client.
func NewStorageLayer(ctx context.Context, credentialsFile string) (*StorageLayer, error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, err
	}
	return &StorageLayer{client: client}, nil
}

// GenerateSignedUploadURL creates a signed URL for uploading an object to GCS.
func (s *StorageLayer) GenerateSignedUploadURL(bucketName, objectPath, contentType string, validFor time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Method:      "PUT",
		ContentType: contentType,
		Expires:     time.Now().Add(validFor),
	}

	url, err := s.client.Bucket(bucketName).SignedURL(objectPath, opts)
	if err != nil {
		return "", err
	}

	return url, nil
}

// InitiateEulaUpload is the resolver for the initiateEulaUpload field.
func (r *mutationResolver) InitiateEulaUpload(ctx context.Context, version int, content string) (*model.PreSignedURLOutput, error) {
	// Authenticate the user
	/*loggedInUser, _ := r.Authenticate.AuthenticateGQLContext(&ctx)
	if loggedInUser == nil {
		return nil, errors.New("access denied")
	}*/

	// Generate a unique file path
	filePath := fmt.Sprintf("eulas/%d-%s", version, uuid.New().String())

	// Initialize the StorageLayer (adjust the credentials file path accordingly)
	storageLayer, err := NewStorageLayer(ctx, "/Users/sabest/Downloads/vport-cloud-fbc45aa137d8.json")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage layer: %w", err)
	}

	// Generate the pre-signed URL
	preSignedURL, err := storageLayer.GenerateSignedUploadURL(bucketName, filePath, contentType, validFor)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed URL: %w", err)
	}

	// Save the EULA record to the database
	eula := vclouddb.Eula{
		Version:  version,
		Content:  content,
		FilePath: filePath,
		Status:   vclouddb.Pending,
		//	UserUUID: loggedInUser.User.UUID,
	}
	createdEula, err := r.DB.CreateEula(eula)
	if err != nil {
		return nil, fmt.Errorf("error saving EULA: %w", err)
	}

	return &model.PreSignedURLOutput{
		URL:      preSignedURL,
		FilePath: createdEula.FilePath,
	}, nil
}

// InitiateEulaUpload is the resolver for the initiateEulaUpload field.
/*func (r *mutationResolver) InitiateEulaUpload(ctx context.Context, version int, content string) (*model.PreSignedURLOutput, error) {
// Authenticate the user
/*loggedInUser, _ := r.Authenticate.AuthenticateGQLContext(&ctx)
if loggedInUser == nil {
	return nil, errors.New("access denied")
}*/

// Generate a unique file path
/*	filePath := fmt.Sprintf("eulas/%d-%s", version, uuid.New().String())

	// Generate the pre-signed URL (this is a placeholder, you need to implement the actual logic)
	preSignedURL := fmt.Sprintf("%s/%s/%s", gcsAPIURL, bucketName, filePath)

	// Save the EULA record to the database
	eula := vclouddb.Eula{
		Version:  version,
		Content:  content,
		FilePath: filePath,
		Status:   vclouddb.Pending,
		//	UserUUID: loggedInUser.User.UUID,
	}
	// Use the CreateEula method to save the EULA record
	createdEula, err := r.DB.CreateEula(eula)
	if err != nil {
		return nil, fmt.Errorf("error saving EULA: %w", err)
	}

	return &model.PreSignedURLOutput{
		URL:      preSignedURL,
		FilePath: createdEula.FilePath,
	}, nil
}*/

// CompleteEulaUpload is the resolver for the completeEulaUpload field.
func (r *mutationResolver) CompleteEulaUpload(ctx context.Context, version int) (*model.EulaOutput, error) {
	// Authenticate the user
	/*	loggedInUser, _ := r.Authenticate.AuthenticateGQLContext(&ctx)
		if loggedInUser == nil {
			return nil, errors.New("access denied")
		}*/

	// Find the EULA by version and update its status to 'Uploaded'
	eula, err := r.DB.UpdateEulaStatus(version, vclouddb.Uploaded)
	if err != nil {
		return nil, fmt.Errorf("error updating EULA status: %w", err)
	}

	publicURL := fmt.Sprintf("%s/%s/%s", gcsAPIURL, bucketName, eula.FilePath)

	return &model.EulaOutput{
		Version:   eula.Version,
		PublicURL: publicURL,
		Status:    model.MapEulaStatus(eula.Status),
	}, nil
}

// SetEulaAvailability is the resolver for the setEulaAvailability field.
func (r *mutationResolver) SetEulaAvailability(ctx context.Context, version int, available bool) (*model.EulaOutput, error) {
	// Authenticate the user
	loggedInUser, _ := r.Authenticate.AuthenticateGQLContext(&ctx)
	if loggedInUser == nil {
		return nil, errors.New("access denied")
	}

	// Find the EULA by version using helper method
	eula, err := r.DB.FindEulaByVersion(version)
	if err != nil {
		return nil, fmt.Errorf("error finding EULA: %w", err)
	}

	// Determine the new status based on availability
	var newStatus vclouddb.EulaStatus
	if available {
		newStatus = vclouddb.Available
	} else {
		newStatus = vclouddb.Deprecated
	}

	// Update the EULA status
	eula, err = r.DB.UpdateEulaStatus(version, newStatus)
	if err != nil {
		return nil, fmt.Errorf("error updating EULA status: %w", err)
	}

	publicURL := fmt.Sprintf("%s/%s/%s", gcsAPIURL, bucketName, eula.FilePath)

	return &model.EulaOutput{
		Version:   eula.Version,
		PublicURL: publicURL,
		Status:    model.MapEulaStatus(eula.Status),
	}, nil
}
