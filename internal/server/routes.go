package server

import (
	"net/http"
    "io"
    "os"
    "bytes"
    "encoding/json"
    "fmt"
    "strconv"
	"database/sql"
    "errors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
    "resume-backend-parser/internal/models"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", s.HelloWorldHandler)

	e.GET("/health", s.healthHandler)

	e.POST("/signup", s.SignupHandler)
	e.POST("/login", s.LoginHandler)
	e.POST("/uploadResume", s.UploadResumeHandler)
	e.POST("/admin/job", s.CreateJobOpeningHandler)
	e.GET("/admin/job/:job_id", s.AdminGetJobOpeningHandler)
	e.GET("/admin/applicants", s.AdminGetApplicantsHandler)
	e.GET("/admin/applicant/:applicant_id", s.AdminGetApplicantHandler)
	e.GET("/jobs", s.GetJobOpeningsHandler)
	e.POST("/jobs/apply", s.ApplyJobHandler)

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) SignupHandler(c echo.Context) error {
    defer c.Request().Body.Close()

    body, _ := io.ReadAll(c.Request().Body)
    c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

    var apiReq models.SignUpRequest
    err := json.NewDecoder(c.Request().Body).Decode(&apiReq)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }

    userExists, err := s.db.UserExists(apiReq.Email)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    if userExists {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "User already exists"})
    }

    password_encrypted, err := PassToHash(apiReq.Password)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }

    err = s.db.CreateUser(apiReq.Name, apiReq.Email, password_encrypted, apiReq.Address, apiReq.ProfileHeadline)

    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    return c.JSON(http.StatusOK, map[string]string{"message": "User created successfully"})
}

func (s *Server) LoginHandler(c echo.Context) error {
    defer c.Request().Body.Close()

    body, _ := io.ReadAll(c.Request().Body)
    c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

    var apiReq models.LoginRequest
    err := json.NewDecoder(c.Request().Body).Decode(&apiReq)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }

    userExists, err := s.db.UserExists(apiReq.Email)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    if !userExists {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "User does not exist"})
    }

    password_encrypted, err := PassToHash(apiReq.Password)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }

    login, err := s.db.CheckUserLogin(apiReq.Email, password_encrypted)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    if !login {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid credentials"})
    }
    var apiResp models.LoginResponse
    apiResp.Token, err = CreateTokens(apiReq.Email)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    return c.JSON(http.StatusOK, apiResp)
}

func (s *Server) UploadResumeHandler(c echo.Context) error {
    token := c.Request().Header.Get("Authorization")
    if token == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }

    user, err := DecodeAuthToken(token)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }
    isAdmin, err := s.db.IsUserAdmin(user)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    if isAdmin {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Only regular users can upload resumes"})
    }
    id, err := s.db.GetUserId(user)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }

    file, handler, err := c.Request().FormFile("file")
    if err != nil {
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Error retrieving the resume."})
    }
    defer file.Close()

    fmt.Printf("Uploaded File: %+v\n", handler.Filename)
    fmt.Printf("File Size: %+v\n", handler.Size)
    fmt.Printf("MIME Header: %+v\n", handler.Header)

    current_dir, err := os.Getwd()
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }

    destination_folder := current_dir + "/resumes/" + strconv.Itoa(id) + "/"
    // remove directory if it exists
    err = os.RemoveAll(destination_folder)
    if err != nil {
        fmt.Println("Error removing directory:", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    err = os.Mkdir(destination_folder, 0755)
    if err != nil {
        fmt.Println("Error creating directory:", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    dst, err := os.Create(destination_folder + handler.Filename)
    if err != nil {
        fmt.Println("Error Creating the File")
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    defer dst.Close()

    _, err = io.Copy(dst, file)
    if err != nil {
        fmt.Println("Error Saving the File")
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }

    //Create profile
    err = s.db.CreateProfile(id, handler.Filename)


    return c.JSON(http.StatusOK, map[string]string{"message": "Resume uploaded successfully"})
}

func (s *Server) CreateJobOpeningHandler(c echo.Context) error {
    token := c.Request().Header.Get("Authorization")
    if token == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }
    user, err := DecodeAuthToken(token)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }
    isAdmin, err := s.db.IsUserAdmin(user)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    fmt.Println(user)
    if !isAdmin {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }
    userId, err := s.db.GetUserId(user)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }

    defer c.Request().Body.Close()
    body, _ := io.ReadAll(c.Request().Body)
    c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

    var apiReq models.CreateJobRequest
    err = json.NewDecoder(c.Request().Body).Decode(&apiReq)
    if err != nil {
        fmt.Println(apiReq)
        fmt.Println(err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }
    totalApplications, err := strconv.Atoi(apiReq.TotalApplications)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }

    err = s.db.CreateJob(apiReq.Title, apiReq.Description, apiReq.CompanyName, totalApplications, userId)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    return c.JSON(http.StatusOK, map[string]string{"message": "Job created successfully"})
}

func (s *Server) AdminGetJobOpeningHandler(c echo.Context) error {
    token := c.Request().Header.Get("Authorization")
    if token == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }

    user, err := DecodeAuthToken(token)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }
    isAdmin, err := s.db.IsUserAdmin(user)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    if !isAdmin {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }

    job_id := c.Param("job_id")
    fmt.Println(job_id)
    if job_id == "" {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }
    jobId, err := strconv.Atoi(job_id)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }
    var apiResp models.AdminGetJobResponse
    apiResp.Job, err = s.db.GetJob(jobId)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    fmt.Println(apiResp)
    apiResp.Applicants, err = s.db.GetApplicants(jobId)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    fmt.Println(apiResp)
    return c.JSON(http.StatusOK, apiResp)
}

func (s *Server) AdminGetApplicantsHandler(c echo.Context) error {
    token := c.Request().Header.Get("Authorization")
    if token == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }

    user, err := DecodeAuthToken(token)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }
    isAdmin, err := s.db.IsUserAdmin(user)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    if !isAdmin {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }

    var applicants models.ApplicantsResponse
    applicants.Applicants, err = s.db.GetAllApplicants()
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    return c.JSON(http.StatusOK, applicants)
}

func (s *Server) AdminGetApplicantHandler(c echo.Context) error {
    token := c.Request().Header.Get("Authorization")
    if token == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }

    user, err := DecodeAuthToken(token)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }
    isAdmin, err := s.db.IsUserAdmin(user)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }
    if !isAdmin {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }

    applicant_id := c.Param("applicant_id")
    if applicant_id == "" {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }
    fmt.Println(applicant_id)
    applicantId, err := strconv.Atoi(applicant_id)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }
    var apiResp models.ApplicantResponse
    apiResp.Applicant, err = s.db.GetApplicantProfile(applicantId)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, apiResp)
}

func (s *Server) GetJobOpeningsHandler(c echo.Context) error {
    token := c.Request().Header.Get("Authorization")
    if token == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }

    user, err := DecodeAuthToken(token)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }
    fmt.Println("token verified")
    fmt.Println(user)

    var apiResp models.GetJobsResponse

    apiResp.Jobs, err = s.db.GetJobs()
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }

    return c.JSON(http.StatusOK, apiResp)
}

func (s *Server) ApplyJobHandler(c echo.Context) error {
    token := c.Request().Header.Get("Authorization")
    if token == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }

    user, err := DecodeAuthToken(token)
    if err != nil {
        fmt.Println(err)
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
    }
    id, err := s.db.GetUserId(user)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
    }

    var job_id = c.QueryParam("job_id")
    if job_id == "" {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }
    jobId, err := strconv.Atoi(job_id)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }

    err = s.db.ApplyJob(jobId, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return c.JSON(http.StatusBadRequest, map[string]string{"error": "Job does not exist"})
        }
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error applying to job"})
    }

	return c.JSON(http.StatusOK, map[string]string{"message": "Successfully applied to job"})

}
