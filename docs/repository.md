# Repository setup

The supplied scaffold had no Git repository. Local Git is initialized on main. An origin remote is now configured as https://github.com/DavFount/QueryLight.git (observed during setup; not added by this task). Nothing has been pushed and no license is chosen. Confirm its ownership, visibility, authentication, and whether to normalize the hosted repository name to querylight before pushing. Commit lockfiles, generated Wails bindings, and packaging inputs; ignore dependencies/build output/secrets/database data. Regenerate bindings instead of hand-editing them.

Before creating the remote, confirm provider/hostname, account or organization owner, public/private visibility, license (or explicit deferral), and SSH/HTTPS authentication with permission to create/push. Repository name: querylight. Do not infer ownership from the scaffold author email.

After reviewing the working tree:

```powershell
git add .
git diff --cached --stat
git commit -m "chore: establish Querylight foundation"
# Add the confirmed remote URL and push main afterward.
```

