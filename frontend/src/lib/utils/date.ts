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

export function formatDate(timestamp: number): string {
    if (!timestamp) return "";

    const date = new Date(timestamp * 1000);
    if (isNaN(date.getTime())) return "";

    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const year = date.getFullYear();

    return `${month}-${day}-${year}`;
}

export function formatDateWithSeconds(timestamp: number): string {
    if (!timestamp) return "";

    const date = new Date(timestamp * 1000);
    if (isNaN(date.getTime())) return "";
    let hour  = String(date.getHours()).padStart(2, "0");
    if (Number(hour) > 12) {
        hour = String(Number(hour) - 12).padStart(2, "0");
    }
    const min = String(date.getMinutes()).padStart(2, "0");
    const sec = String(date.getSeconds()).padStart(2, "0");
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const year = date.getFullYear();

    return `${month}-${day}-${year} ${hour}:${min}:${sec} ${date.getHours() >= 12 ? "PM" : "AM"}`;
}