package models

// 회원가입 시 리턴 타입
type SignupResult struct {
	Sendmail bool `json:"sendmail"`
	Target   uint `json:"target"`
}

// 인증 완료하기 파라미터
type VerifyParameter struct {
	Target   uint
	Code     string
	Id       string
	Password string
	Name     string
}

// 비밀번호 초기화 시 리턴 타입
type ResetPasswordResult struct {
	Sendmail bool `json:"sendmail"`
}

// 구글 OAuth 응답
type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// 네이버 OAuth 응답
type NaverUser struct {
	Response struct {
		Email        string `json:"email"`
		Nickname     string `json:"nickname"`
		ProfileImage string `json:"profile_image"`
	} `json:"response"`
}

// 카카오 OAuth 응답
type KakaoUser struct {
	ID           int64 `json:"id"`
	KakaoAccount struct {
		Email   string `json:"email"`
		Profile struct {
			Nickname        string `json:"nickname"`
			ProfileImageUrl string `json:"profile_image_url"`
		} `json:"profile"`
	} `json:"kakao_account"`
}
