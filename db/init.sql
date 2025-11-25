CREATE TYPE merge_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE Users (
  user_id VARCHAR PRIMARY KEY,
  username VARCHAR UNIQUE,
  team_name VARCHAR,
  is_active BOOLEAN
);

CREATE TABLE Teams (
  team_name VARCHAR PRIMARY KEY,
  members JSONB
);

CREATE TABLE Pull_requests (
  pull_request_id VARCHAR PRIMARY KEY,
  pull_request_name VARCHAR,
  author_id VARCHAR,
  status merge_status,
  assigned_reviewers VARCHAR[],
  created_at TIMESTAMP,
  merged_at TIMESTAMP,
  FOREIGN KEY (author_id) REFERENCES Users (user_id)
);
