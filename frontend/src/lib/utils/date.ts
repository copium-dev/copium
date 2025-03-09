export function offsetTimezone(timestamp: number) {
    if (!timestamp) return 0;

    const date = new Date(timestamp);
    if (isNaN(date.getTime())) return 0;

    // adjust for timezone
    const adjustedDate = new Date(
        date.getTime() + date.getTimezoneOffset() * 60 * 1000
    );

    // return as unix timestamp in seconds
    return Math.floor(adjustedDate.getTime() / 1000); 
}