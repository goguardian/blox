
#### Summary
<!-- What does this pull request do? -->

#### Implementation details
<!-- How are the changes implemented? -->

#### Testing
<!-- How was this tested? -->
- [ ] cluster-state-service binary built locally and unit-tests pass (`cd cluster-state-service; make; cd ../`)
- [ ] cluster-state-service build in Docker succeeds (`cd cluster-state-service; make release; cd ../`)
- [ ] daemon-scheduler binary built locally and unit-tests pass (`cd daemon-scheduler; make; cd ../`)
- [ ] daemon-scheduler build in Docker succeeds (`cd daemon-scheduler; make release; cd ../`)

New tests cover the changes: <!-- yes|no -->

#### Description for the changelog
<!--
Write a short summary that describes the changes in this pull request
for inclusion in changelog.
-->

#### Licensing

This contribution is under the terms of the Apache 2.0 License: Yes

#### Before merging
<!-- Run integration and end-to-end tests before merging -->
- [ ] cluster-state-service end-to-end tests pass. Required setup details are listed [here](https://github.com/goguardian/blox/blob/dev/cluster-state-service/internal/Readme.md).
- [ ] daemon-scheduler integration tests pass. Required setup details are listed [here](https://github.com/goguardian/blox/blob/dev/daemon-scheduler/internal/features/README.md).
- [ ] daemon-scheduler end-to-end tests pass. Required setup details are listed [here](https://github.com/goguardian/blox/blob/dev/daemon-scheduler/internal/features/README.md).
