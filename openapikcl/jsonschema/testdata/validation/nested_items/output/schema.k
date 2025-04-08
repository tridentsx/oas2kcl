schema DynatraceMonitoring_environmentGroups_items_environments_items:
    name?: str
    url?: DynatraceMonitoring_environmentGroups_items_environments_items_url


schema DynatraceMonitoring_environmentGroups_items_environments_items_url:
    value?: str
    type?: str


schema DynatraceMonitoring_projects_items:
    name: str
    type?: str
    path?: str


schema DynatraceMonitoring:
    environmentGroups: [DynatraceMonitoring_environmentGroups_items]
    manifestVersion: str
    projects: [DynatraceMonitoring_projects_items]


schema DynatraceMonitoring_environmentGroups_items:
    name?: str
    environments?: [DynatraceMonitoring_environmentGroups_items_environments_items]


