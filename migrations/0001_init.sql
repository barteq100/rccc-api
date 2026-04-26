BEGIN;

CREATE TYPE application_status AS ENUM (
  'saved',
  'applied',
  'interview',
  'offer',
  'rejected'
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE jobs (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  company TEXT NOT NULL,
  location TEXT NOT NULL,
  remote BOOLEAN NOT NULL DEFAULT FALSE,
  description TEXT NOT NULL,
  source TEXT NOT NULL,
  source_url TEXT NOT NULL,
  posted_at TIMESTAMPTZ NOT NULL,
  ingested_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT jobs_title_non_empty CHECK (btrim(title) <> ''),
  CONSTRAINT jobs_company_non_empty CHECK (btrim(company) <> ''),
  CONSTRAINT jobs_source_non_empty CHECK (btrim(source) <> ''),
  CONSTRAINT jobs_source_url_non_empty CHECK (btrim(source_url) <> '')
);

CREATE UNIQUE INDEX idx_jobs_source_source_url ON jobs (source, source_url);
CREATE INDEX idx_jobs_remote_posted_at ON jobs (remote, posted_at DESC);
CREATE INDEX idx_jobs_source_posted_at ON jobs (source, posted_at DESC);

CREATE TRIGGER trg_jobs_set_updated_at
BEFORE UPDATE ON jobs
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE profiles (
  id SMALLINT PRIMARY KEY DEFAULT 1,
  preferred_stack TEXT[] NOT NULL DEFAULT '{}',
  remote_only BOOLEAN NOT NULL DEFAULT FALSE,
  preferred_locations TEXT[] NOT NULL DEFAULT '{}',
  target_seniority TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT profiles_single_row CHECK (id = 1)
);

CREATE TRIGGER trg_profiles_set_updated_at
BEFORE UPDATE ON profiles
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE job_applications (
  job_id TEXT PRIMARY KEY REFERENCES jobs(id) ON DELETE CASCADE,
  status application_status NOT NULL DEFAULT 'saved',
  saved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  applied_at TIMESTAMPTZ,
  status_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT job_applications_applied_at_required
    CHECK (status = 'saved' OR applied_at IS NOT NULL)
);

CREATE INDEX idx_job_applications_status ON job_applications (status, status_changed_at DESC);

CREATE TRIGGER trg_job_applications_set_updated_at
BEFORE UPDATE ON job_applications
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

INSERT INTO profiles (id) VALUES (1)
ON CONFLICT (id) DO NOTHING;

COMMIT;
