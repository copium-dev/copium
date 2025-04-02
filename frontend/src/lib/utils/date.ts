/**
 * IF YOU WERE WONDERING ARE WE MILLISECONDS OR SECONDS...
 * BACKEND STORES MILLISECONDS BECAUSE IT'S KEYSET PAGINATION BY APPLIED DATE
 * SO IT NEEDS TO BE MORE PRECISE THAN SECONDS. HOWEVER, FRONTEND SHOULDN'T BE 
 * DISPLAYING THE SECONDS SO WE CONVERT IT TO SECONDS... OKAY???????? FURTHERMORE,
 * ALL BACKEND RETURNS TO FRONTEND IS IN SECONDS TO FURTHER CONFUSE YOU :D
 */

// for input type="date"
export function formatDateForInput(timestamp: number): string {
    if (!timestamp) return "";

    const date = new Date();
    if (isNaN(date.getTime())) return "";

    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const year = date.getFullYear();

    return `${year}-${month}-${day}`;
}

// for displaying in components
export function formatDateForDisplay(timestamp: number): string {
    if (!timestamp) return "";

    const date = new Date(timestamp);
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

export function convertLocalDateToTimestamp(dateString: string): number {
    // this is just for current time 
    const now = new Date();
    // dateString comes as yyyy-mm-dd
    const [year, month, day] = dateString.split('-').map(Number);

    const [hours, mins, secs, ms] = [now.getHours(), now.getMinutes(), now.getSeconds(), now.getMilliseconds()];
    
    // month is 0-indexed in Date constructor
    const date = new Date(year, month - 1, day, hours, mins, secs, ms);

    // convert to unix timestamp in milliseconds
    return date.getTime();
}