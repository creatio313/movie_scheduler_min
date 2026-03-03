package validators

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/creatio313/movie_scheduler/internal/models"
)

// ValidationError はバリデーションエラーを表します
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateProject はプロジェクトの入力値を検証します
func ValidateProject(p models.Project) error {
	if p.Title == "" {
		return ValidationError{Field: "title", Message: "プロジェクト名は必須です"}
	}
	if utf8.RuneCountInString(p.Title) > 255 {
		return ValidationError{Field: "title", Message: "プロジェクト名は255文字以内で入力してください"}
	}
	return nil
}

// ValidateCast はキャストの入力値を検証します
func ValidateCast(c models.Cast) error {
	if c.Name == "" {
		return ValidationError{Field: "name", Message: "役者名は必須です"}
	}
	if utf8.RuneCountInString(c.Name) > 100 {
		return ValidationError{Field: "name", Message: "役者名は100文字以内で入力してください"}
	}
	if c.RoleName == "" {
		return ValidationError{Field: "role_name", Message: "役名は必須です"}
	}
	if utf8.RuneCountInString(c.RoleName) > 100 {
		return ValidationError{Field: "role_name", Message: "役名は100文字以内で入力してください"}
	}
	return nil
}

// ValidateScene はシーンの入力値を検証します
func ValidateScene(s models.Scene) error {
	if s.SceneName == "" {
		return ValidationError{Field: "scene_name", Message: "シーン名は必須です"}
	}
	if utf8.RuneCountInString(s.SceneName) > 20 {
		return ValidationError{Field: "scene_name", Message: "シーン名は20文字以内で入力してください"}
	}
	return nil
}

// ValidateCandidateDate は候補日の入力値を検証します
func ValidateCandidateDate(cd models.CandidateDate) error {
	if cd.TargetDate == "" {
		return ValidationError{Field: "target_date", Message: "候補日は必須です"}
	}
	// YYYY-MM-DD形式の検証
	_, err := time.Parse("2006-01-02", cd.TargetDate)
	if err != nil {
		return ValidationError{Field: "target_date", Message: "候補日はYYYY-MM-DD形式で入力してください"}
	}
	return nil
}

// ValidateTimeSlotDef は時間枠の入力値を検証します
func ValidateTimeSlotDef(t models.TimeSlotDef) error {
	if t.SlotName == "" {
		return ValidationError{Field: "slot_name", Message: "時間枠名は必須です"}
	}
	if utf8.RuneCountInString(t.SlotName) > 50 {
		return ValidationError{Field: "slot_name", Message: "時間枠名は50文字以内で入力してください"}
	}
	// start_time/end_timeが指定されている場合、HH:MM:SS形式を検証
	if t.StartTime != "" {
		_, err := time.Parse("15:04:05", t.StartTime)
		if err != nil {
			return ValidationError{Field: "start_time", Message: "開始時刻はHH:MM:SS形式で入力してください"}
		}
	}
	if t.EndTime != "" {
		_, err := time.Parse("15:04:05", t.EndTime)
		if err != nil {
			return ValidationError{Field: "end_time", Message: "終了時刻はHH:MM:SS形式で入力してください"}
		}
	}
	return nil
}
