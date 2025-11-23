CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    reviewer_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer_id ON pr_reviewers(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pull_request_id ON pr_reviewers(pull_request_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_assigned_at ON pr_reviewers(assigned_at);