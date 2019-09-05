package mosolyapi

import (
	"context"
	"time"
)

// ClientMock is an Mosoly mock API client.
type ClientMock struct{}

// GetUserUpdates is a mock for Mosoly API user/mentor updates
func (c *ClientMock) GetUserUpdates(ctx context.Context, since time.Time) ([]User, error) {
	return []User{
		{
			ID:            1,
			InviteURLHash: "ea03d482a5d9a536dc3f0f108ca543c1a7179d51296a0ece50a447e512b06d77",
			Account:       "0x690e4721ca6da17c9e66c6b988e6b35635e6ec3b",
			Validated:     false,
			JoinedAt:      time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			CreatedAt:     time.Now().UTC(),
			Mentorees: []Mentoree{
				{
					UserID:              3,
					CustomNameForMentor: "",
					Account:             "0x11111220f57c8e7e3a45a415afba94b2ae6dc16e",
					Validated:           false,
					MentorshipStarted:   time.Now().UTC(),
				},
				{
					UserID:              4,
					CustomNameForMentor: "",
					Account:             "0x00000220f57c8e7e3a45a415afba94b2ae6dc16e",
					Validated:           false,
					MentorshipStarted:   time.Now().UTC(),
				},
			},
		},
		{
			ID:            2,
			InviteURLHash: "90ab00d89580b62b2850eb4c1ee0d5ce0b7231e8af91a69f4da134f55d0bc4f1",
			Account:       "0x11111220f57c8e7e3a45a415afba94b2ae6dc16e",
			Validated:     false,
			JoinedAt:      time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			CreatedAt:     time.Now().UTC(),
			Mentorees: []Mentoree{
				{
					UserID:              3,
					CustomNameForMentor: "",
					Account:             "0x11111220f57c8e7e3a45a415afba94b2ae6dc16e",
					Validated:           false,
					MentorshipStarted:   time.Now().UTC(),
				},
				{
					UserID:              4,
					CustomNameForMentor: "",
					Account:             "0x00000220f57c8e7e3a45a415afba94b2ae6dc16e",
					Validated:           false,
					MentorshipStarted:   time.Now().UTC(),
				},
			},
		},
		{
			ID:            3,
			InviteURLHash: "0d326a1fdc19809f7d221164e8ffda7c2ba9d8622260aab34c7f45a63b146926",
			Account:       "0x00000220f57c8e7e3a45a415afba94b2ae6dc16e",
			Validated:     false,
			JoinedAt:      time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			CreatedAt:     time.Now().UTC(),
			Mentorees:     nil,
			Mentors: []Mentor{
				{
					UserID:            1,
					Account:           "0x690e4721ca6da17c9e66c6b988e6b35635e6ec3b",
					MentorshipStarted: time.Now().UTC(),
				},
				{
					UserID:            2,
					Account:           "0x11111220f57c8e7e3a45a415afba94b2ae6dc16e",
					MentorshipStarted: time.Now().UTC(),
				},
			},
		},
		{
			ID:            4,
			InviteURLHash: "ceebf77a833b30520287ddd9478ff51abbdffa30aa90a8d655dba0e8a79ce0c1",
			Account:       "0x001a38800771fb9b27678d46078231936f763250",
			Validated:     false,
			JoinedAt:      time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			CreatedAt:     time.Now().UTC(),
			Mentorees: []Mentoree{
				{
					UserID:              5,
					CustomNameForMentor: "",
					Account:             "0x17C717Fe53A5AF967Cc16AbfF237A9275c161948",
					Validated:           false,
					MentorshipStarted:   time.Now().UTC(),
				},
			},
			Mentors: []Mentor{
				{
					UserID:            1,
					Account:           "0x690e4721ca6da17c9e66c6b988e6b35635e6ec3b",
					MentorshipStarted: time.Now().UTC(),
				},
				{
					UserID:            2,
					Account:           "0x11111220f57c8e7e3a45a415afba94b2ae6dc16e",
					MentorshipStarted: time.Now().UTC(),
				},
			},
		},
		{
			ID:            5,
			InviteURLHash: "e455bf8ea6e7463a1046a0b52804526e119b4bf5136279614e0b1e8e296a4e2d",
			Account:       "0x17C717Fe53A5AF967Cc16AbfF237A9275c161948",
			Validated:     false,
			JoinedAt:      time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			CreatedAt:     time.Now().UTC(),
			Mentorees:     nil,
			Mentors: []Mentor{
				{
					UserID:            4,
					Account:           "0x001a38800771fb9b27678d46078231936f763250",
					MentorshipStarted: time.Now().UTC(),
				},
			},
		},
	}, nil
}

// GetProjectUpdates is a mock for Mosoly API project updates
func (c *ClientMock) GetProjectUpdates(ctx context.Context, since time.Time) ([]Project, error) {
	// TODO: Implement mock object to be retrieved from Projects
	return []Project{}, nil
}
