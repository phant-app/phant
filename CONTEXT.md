# Phant

Phant manages PHP and Laravel development environments from a terminal. Its language defines the product boundaries independently of its terminal implementation.

## Language

**Phant Environment**:
One host machine where Phant runs, whether accessed locally or through SSH. Its PHP runtimes, services, sites, configuration, and dump collector belong to that host.
_Avoid_: Project environment, managed environment

**Phant Terminal Session**:
One interactive Phant process connected to a Phant Environment. The session owns the running dump collector while users move between terminal views.
_Avoid_: Dump page session, collector daemon

**Session Dump Buffer**:
A bounded in-memory collection of dump events owned by one Phant Terminal Session. Its events are discarded when that session ends.
_Avoid_: Dump history, persistent dump store

**Managed Change**:
A user-approved adjustment to a Phant Environment that Phant previews before attempting to apply it automatically while showing the intended and executed commands. If automatic application fails, Phant provides the equivalent manual command.
_Avoid_: Silent mutation, opaque automation

Viewing or refreshing Phant Environment information is not a Managed Change and must not alter the environment.

**Phant-Owned Configuration**:
An isolated configuration file or fragment created and managed by Phant. Phant prefers these over modifying configuration maintained by users or other tools.
_Avoid_: User configuration, shared configuration

**Project Context**:
An optional filter within a Phant Environment that narrows project-related information such as sites and dump events. It does not own host-wide PHP runtimes, services, or configuration.
_Avoid_: Project environment, active environment
