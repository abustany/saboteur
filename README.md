# Saboteur 💣

GitHub merge bot that merges branch fast-forward, using Git.

## Why?

There is a lot of merge bots for GitHub, but all the ones I checked merge pull
requests using the GitHub API. This leaves you with two options:

1. Let the GitHub API merge your PRs using merge commits, not ideal if you
   prefer linear histories

2. Let the GitHub API merge your PRs using the "squash" or "rebase" methods,
   which both modify your commits, not ideal if you like your GPG signatures to
   be preserved

Saboteur will only merge PRs that are fast forward (rebased on top of) their
base branch, and will do so using a regular `git push`, that preserves commit
SHAs.

## Merge rules

Saboteur can select the pull requests to merge based on the following criteria:

- Target branch
- Successful checks (eg. GitHub Actions results)
- Labels

## Setup

Saboteur can authenticate itself using either personal access tokens or as a
GitHub app.

### Personal access tokens

Create a [classic Personal Access
Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
with the `repo` scope.

**Note: fine grained access tokens are currently unsupported, since they don't
work well with the GraphQL API used by Saboteur**.

Check the examples in [saboteur.yml](saboteur.yml) to see how to set the token
in the config file, either directly or via an environment variable.

### GitHub apps

[Register a new GitHub
app](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app),
setting neither a callback URL nor a webhook URL. Set the repository
permissions as follows:

- Actions: read only
- Checks: read only
- Commit statuses: read only
- Content: read and write
- Pull requests: read only

Leave organization permissions and account permissions empty.

Once the app is created, write down the application ID on top of the page. Go
to the bottom of the page and generate a new private key for the app. Go to the
"Install App" section, and install the app into your account. Once the app is
installed, retrieve the installation ID from the URL: if the URL looks like
`https://github.com/organizations/my-org/settings/installations/12345678`, the
installation ID is `12345678`.

Check the example in [saboteur.yml](saboteur.yml) to see where to input those
values.

## Is this project ready/stable?

This project is at the MVP stage: you can call it from a cron job, and it'll
merge PRs as instructed. The damages it can cause should be quite limited,
since it only attempts fast forward pushes.

Potential future improvements include:
- listening to webhook events for faster reactions
- figure out if cloning is really necessary?
- sandboxing the git operations better, whether it is using libgit2's SQLite
  ODB backend, a WASM build of libgit2…

## Isn't cloning a repo for each merge expensive?

According to my understanding of git, one should be able to do a `git push`
without cloning at all (we are just asking Git to update a remote ref in a fast
forward way). This doesn't seem to work in practice for reasons I don't fully
understand, so Saboteur will instead do a shallow clone of your repo, using an
object filter to filter out blobs (file contents). What this means is that
cloning is pretty fast, eg. cloning the latest master of torvalds/linux with
this method downloads less than 5MB. When GitHub implements support for
additional object filters, we can make this process even more efficient (we
currently clone tree objects, which we don't need)
