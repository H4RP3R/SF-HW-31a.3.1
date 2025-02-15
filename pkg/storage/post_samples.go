package storage

var TestPosts = []Post{
	{
		ID:          1,
		Title:       "Post 1",
		Content:     "This is the content of post 1",
		AuthorID:    1,
		AuthorName:  "Mark",
		CreatedAt:   1643723400, // 2022-02-01 12:00:00
		PublishedAt: 1643723400, // 2022-02-01 12:00:00
	},
	{
		ID:          2,
		Title:       "Post 2",
		Content:     "This is the content of post 2",
		AuthorID:    2,
		AuthorName:  "Tom",
		CreatedAt:   1643809800, // 2022-02-02 12:00:00
		PublishedAt: 1643809800, // 2022-02-02 12:00:00
	},
	{
		ID:          3,
		Title:       "Post 3",
		Content:     "This is the content of post 3",
		AuthorID:    1,
		AuthorName:  "Mark",
		CreatedAt:   1643896200, // 2022-02-03 12:00:00
		PublishedAt: 1643896200, // 2022-02-03 12:00:00
	},
	{
		ID:          4,
		Title:       "Post 4",
		Content:     "This is the content of post 4",
		AuthorID:    3,
		AuthorName:  "Travis",
		CreatedAt:   1643982600, // 2022-02-04 12:00:00
		PublishedAt: 1643982600, // 2022-02-04 12:00:00
	},
	{
		ID:          5,
		Title:       "Post 5",
		Content:     "This is the content of post 5",
		AuthorID:    2,
		AuthorName:  "Tom",
		CreatedAt:   1644069000, // 2022-02-05 12:00:00
		PublishedAt: 1644069000, // 2022-02-05 12:00:00
	},
}
