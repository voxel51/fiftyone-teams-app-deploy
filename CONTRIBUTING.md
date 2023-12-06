# Contributing

Our Helm Chart's README.md is automatically
generated using the pre-commit hooks for

* [https://github.com/norwoodj/helm-docs](https://github.com/norwoodj/helm-docs)
* [https://github.com/Lucas-C/pre-commit-hooks-nodejs](https://github.com/Lucas-C/pre-commit-hooks-nodejs)

1. Install dependent tools
    * Using [asdf](https://asdf-vm.com/)

        ```shell
        make asdf
        ```

    * Manually
      * [helm-docs](https://github.com/norwoodj/helm-docs#installation)
      * [pre-commit](https://pre-commit.com/#installation)

1. Install the pre-commit hooks

    ```shell
    make hooks
    ```

1. Update the Go Template
  [helm/fiftyone-teams-app/README.md.gotmpl](./helm/fiftyone-teams-app/README.md.gotmpl).
1. To render
  [helm/fiftyone-teams-app/README.md](./helm/fiftyone-teams-app/README.md)
    * Add the changed file `helm/fiftyone-teams-app/README.md.gotmpl`
    * Either
      * Commit the changes and let the hooks render from the template

          ```shell
          [fiftyone-teams-app-deploy]$ git add helm/fiftyone-teams-app/README.md.gotmpl
          [fiftyone-teams-app-deploy]$ git commit -m 'adding new section'
          check for added large files...........................................Passed
          check for case conflicts..............................................Passed
          check that scripts with shebangs are executable.......................Passed
          check yaml........................................(no files to check)Skipped
          detect aws credentials................................................Passed
          fix end of files......................................................Passed
          mixed line ending.....................................................Passed
          pretty format json................................(no files to check)Skipped
          trim trailing whitespace..............................................Passed
          No-tabs checker.......................................................Passed
          markdownlint......................................(no files to check)Skipped
          markdownlint-fix..................................(no files to check)Skipped
          codespell.............................................................Passed
          yamllint..........................................(no files to check)Skipped
          Helm Docs.............................................................Failed
          - hook id: helm-docs
          - files were modified by this hook

          INFO[2023-11-09T16:11:14-07:00] Found Chart directories [.]
          INFO[2023-11-09T16:11:14-07:00] Generating README Documentation for chart helm/fiftyone-teams-app

          Insert a table of contents in Markdown files, like a README.md........Passed
          [fiftyone-teams-app-deploy]$ git add helm/fiftyone-teams-app/README.md
          [fiftyone-teams-app-deploy]$ git commit -m 'adding new section'
          check for added large files...........................................Passed
          check for case conflicts..............................................Passed
          check that scripts with shebangs are executable.......................Passed
          check yaml........................................(no files to check)Skipped
          detect aws credentials................................................Passed
          fix end of files......................................................Passed
          mixed line ending.....................................................Passed
          pretty format json................................(no files to check)Skipped
          trim trailing whitespace..............................................Passed
          No-tabs checker.......................................................Passed
          markdownlint..........................................................Passed
          markdownlint-fix......................................................Passed
          codespell.............................................................Passed
          yamllint..........................................(no files to check)Skipped
          Helm Docs.............................................................Passed
          Insert a table of contents in Markdown files, like a README.md.......................Passed
          [AS-22-helm-docs a81c21b] adding new section
          2 files changed, 10 insertions(+)
          ```

      * Manually run the pre-commit hooks

          ```shell
          git add helm/fiftyone-teams-app/README.md.gotmpl
          pre-commit run helm-docs
          pre-commit run markdown-toc
          git add helm/fiftyone-teams-app/README.md
          git commit -m '<COMMIT_MESSAGE>'
          ```
