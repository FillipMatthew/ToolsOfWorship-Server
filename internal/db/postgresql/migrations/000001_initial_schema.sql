CREATE TABLE IF NOT EXISTS SignKeys (
    id UUID PRIMARY KEY,
    key BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS EncKeys (
    id UUID PRIMARY KEY,
    key BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS Users (
    id UUID PRIMARY KEY,
    displayName VARCHAR(50) NOT NULL,
    created TIMESTAMPTZ NOT NULL,
    isDeleted BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE TABLE IF NOT EXISTS UserConnections (
    userId UUID NOT NULL REFERENCES Users(id),
    signInType INTEGER NOT NULL,
    accountId TEXT NOT NULL,
    authDetails TEXT,
    PRIMARY KEY (userId, signInType),
    UNIQUE (signInType, accountId)
);

CREATE TABLE IF NOT EXISTS Fellowships (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    creator UUID NOT NULL REFERENCES Users(id)
);

CREATE TABLE IF NOT EXISTS FellowshipMembers (
    fellowshipId UUID REFERENCES Fellowships(id),
    userId UUID REFERENCES Users(id),
    access INTEGER NOT NULL,
    PRIMARY KEY (fellowshipId, userId)
);

CREATE TABLE IF NOT EXISTS FellowshipCircles (
    id UUID PRIMARY KEY,
    fellowshipId UUID NOT NULL REFERENCES Fellowships(id),
    name TEXT NOT NULL,
    type INTEGER NOT NULL,
    creator UUID NOT NULL REFERENCES Users(id)
);

CREATE TABLE IF NOT EXISTS CircleMembers (
    circleId UUID REFERENCES FellowshipCircles(id),
    userId UUID REFERENCES Users(id),
    access INTEGER NOT NULL,
    PRIMARY KEY (circleId, userId)
);

CREATE TABLE IF NOT EXISTS Posts (
    id UUID PRIMARY KEY,
    authorId UUID NOT NULL REFERENCES Users(id),
    fellowshipId UUID,
    circleId UUID,
    posted TIMESTAMPTZ NOT NULL,
    heading VARCHAR(80),
    article TEXT
);

CREATE INDEX IF NOT EXISTS idx_posts_posted ON Posts(posted);
CREATE INDEX IF NOT EXISTS idx_userconnections_accountid ON UserConnections(accountId);
CREATE INDEX IF NOT EXISTS idx_userconnections_userid ON UserConnections(userId);
CREATE INDEX IF NOT EXISTS idx_fellowshipmembers_userid ON FellowshipMembers(userId);
CREATE INDEX IF NOT EXISTS idx_circlemembers_userid ON CircleMembers(userId);
CREATE INDEX IF NOT EXISTS idx_posts_fellowshipid ON Posts(fellowshipId);
CREATE INDEX IF NOT EXISTS idx_posts_circleid ON Posts(circleId);
