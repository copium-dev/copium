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

// for input type="date"
export function formatDateForInput(timestamp: number): string {
    if (!timestamp) return "";

    const date = new Date(timestamp * 1000);
    if (isNaN(date.getTime())) return "";

    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const year = date.getFullYear();

    return `${year}-${month}-${day}`;
}

// for displaying in components
export function formatDateForDisplay(timestamp: number): string {
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

// convert to unix seconds for backend
export function convertLocalDateToTimestamp(dateString: string): number {
    // dateString comes as yyyy-mm-dd
    const [year, month, day] = dateString.split('-').map(Number);
    
    // month is 0-indexed in Date constructor
    const date = new Date(year, month - 1, day, 12, 0, 0);
    
    // return unix timestamp in seconds
    return Math.floor(date.getTime() / 1000);
}