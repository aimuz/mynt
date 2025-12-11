export function getHealthBadgeColor(health?: string): string {
    switch (health?.toUpperCase()) {
        case "ONLINE":
            return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
        case "DEGRADED":
            return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
        case "OFFLINE":
        case "UNAVAIL":
        case "FAULTED":
            return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
        default:
            return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400";
    }
}
