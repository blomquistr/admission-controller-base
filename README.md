# Admission Controller Base

A project to learn some Go by migrating our Python mutating admission controller to Go.

## Functionality to Replicate

Tracking each feature we have implemented in Python that we need to implement in the new controller:

### Webhooks

- [ ] Tolerate Azure spot instances
- [X] Add the internal-only annotation for the cloud provider we're running in
- [X] Reject objects in the default namespace
- [ ] Prevent volumes that use `EmptyDir` storage
- [ ] Reject Service objects that violate CVE-2020-8554 until the core Kubernetes offering does
- [ ] Add DNS operator custom resources for appropriately-annotated services

### Other Features

- [ ] Add new webhooks via code generation and a plugin architecture
- [ ] Add endpoints to enable and disable all webhooks handled by the server
- [ ] Alerting to external communications channels in addition to the response to `kubectl`
  - [ ] Slack (implemented in our admission controller in Python)

### Net-new Improvements

- [ ] A Helm chart for deployment
- [X] Deploy with Kustomize instead of in-house templating logic
- [ ] BATS tests baked into the deployment via a CronJob and a BATS test script
