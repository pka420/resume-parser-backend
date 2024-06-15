package models

import (
	"time"
)

// enum of UserType
type UserType string

const (
	Applicant UserType = "Applicant"
	Admin     UserType = "Admin"
)

type User struct {
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	Address         string   `json:"address"`
	UserType        UserType `json:"userType"`
	PasswordHash    string   `json:"passwordHash"`
	ProfileHeadline string   `json:"profileHeadline"`
	Profile         Profile  `json:"profile"`
}

type Profile struct {
	Applicant         string   `json:"applicant"`
	ResumeFileAddress string `json:"resumeFileAddress"`
	Skills            string `json:"skills"`
	Education         string `json:"education"`
	Experience        string `json:"experience"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
}

type Job struct {
    Id               int       `json:"id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	TotalApplications int       `json:"totalApplications"`
	CompanyName       string    `json:"companyName"`
    Applicants        []int     `json:"applicants"`
    PostedOn          time.Time `json:"postedOn"`
	PostedBy          string      `json:"postedBy"`
}

type SignUpRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
    Address  string `json:"address"`
    ProfileHeadline string `json:"profileHeadline"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token string `json:"token"`
}

type CreateJobRequest struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    CompanyName string `json:"companyName"`
    TotalApplications string `json:"totalApplications"`
}

type CreateJobResponse struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    CompanyName string `json:"companyName"`
    TotalApplications int `json:"totalApplications"`
    NumApplications int `json:"numApplications"`
    PostedOn    time.Time `json:"postedOn"`
    PostedBy    string `json:"postedBy"`
}

type GetJobsResponse struct {
    Jobs []Job `json:"jobs"`
}

type CreateApplicantProfileRequest struct {
    ResumeFileAddress string `json:"resumeFileAddress"`
    Skills            string `json:"skills"`
    Education         string `json:"education"`
    Experience        string `json:"experience"`
    Name              string `json:"name"`
    Email             string `json:"email"`
    Address           string `json:"address"`
}


type AdminGetJobResponse struct {
    Job Job `json:"job"`
    Applicants []Profile `json:"applicants"`
}

type ApplicantsResponse struct {
    Applicants []Profile `json:"applicants"`
}

type ApplicantResponse struct {
    Applicant Profile `json:"applicant"`
}


type ProfileThirdParty struct {
    Name       string     `json:"name"`
    Email      string     `json:"email"`
    Phone      string     `json:"phone"`
    Education  []Institute `json:"education"`
    Experience []Experience `json:"experience"`
    Skills     []string   `json:"skills"`
}

type Institute struct {
    Name string `json:"name"`
}

type Experience struct {
    Role string `json:"role"`
}
