# Remote-native terminal execution

Phant v2.0 runs on and manages the host where its binary executes, including a host reached through SSH. It does not include a local controller for managing remote hosts, because direct host-local execution reuses ordinary permissions and avoids an SSH control plane, credential management, and agent protocol.
