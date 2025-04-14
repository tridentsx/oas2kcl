schema DynatraceMonitoring_environmentGroups_items:
    name?: str
    environments?: [DynatraceMonitoring_environmentGroups_items_environments_items]


schema DynatraceMonitoring_environmentGroups_items_environments_items:
    name?: str
    url?: DynatraceMonitoring_environmentGroups_items_environments_items_url


schema DynatraceMonitoring_environmentGroups_items_environments_items_url:
    type?: str
    value?: str


schema DynatraceMonitoring:
    projects: [DynatraceMonitoring_projects_items]
    environmentGroups: [DynatraceMonitoring_environmentGroups_items]
    manifestVersion: str


schema DynatraceMonitoring_projects_items:
    path?: str
    name: str
    type?: str


