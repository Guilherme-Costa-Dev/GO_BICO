package domain

type User struct {
	ID    string `json:"id" firestore:"-"`
	Nome  string `json:"nome" firestore:"nome"`
	Email string `json:"email" firestore:"email"`
}
