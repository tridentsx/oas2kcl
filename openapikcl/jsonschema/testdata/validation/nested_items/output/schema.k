schema DynatraceMonitoring:
    environmentGroups: [DynatraceMonitoring_environmentGroups_items]
    manifestVersion: str
    projects: [DynatraceMonitoring_projects_items]


schema DynatraceMonitoring_environmentGroups_items:
    environments?: [DynatraceMonitoring_environmentGroups_items_environments_items]
    name?: str


schema DynatraceMonitoring_environmentGroups_items_environments_items:
    name?: str
    url?: DynatraceMonitoring_environmentGroups_items_environments_items_url


schema DynatraceMonitoring_environmentGroups_items_environments_items_url:
    type?: str
    value?: str


schema DynatraceMonitoring_projects_items:
    name: str
    type?: str
    path?: str


