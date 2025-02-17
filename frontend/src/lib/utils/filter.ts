export function buildParamsFromFilters({
    query,
    company,
    role,
    location,
    startDate,
    endDate,
    status,
}: {
    query: string;
    company: string;
    role: string;
    location: string;
    startDate: string;
    endDate: string;
    status: string;
}): URLSearchParams {
    const params = new URLSearchParams(window.location.search);

    if (query) {
        params.set("q", query);
    } else {
        params.delete("q");
    }
    if (company) {
        params.set("company", company);
    } else {
        params.delete("company");
    }
    if (role) {
        params.set("role", role);
    } else {
        params.delete("role");
    }
    if (location) {
        params.set("location", location);
    } else {
        params.delete("location");
    }
    if (status && status !== "Status") {
        params.set("status", status);
    } else {
        params.delete("status");
    }
    if (startDate) {
        const startTimestamp = new Date(startDate).getTime();
        params.set("startDate", startTimestamp.toString());
    } else {
        params.delete("startDate");
    }
    if (endDate) {
        const endTimestamp = new Date(endDate).getTime();
        params.set("endDate", endTimestamp.toString());
    } else {
        params.delete("endDate");
    }

    return params;
}

export function changePage(direction: "next" | "prev" | number): URLSearchParams {
    const params = new URLSearchParams(window.location.search);
    const currentPage = parseInt(params.get("page") || "0", 10);

    if (typeof direction === "number") {
        // algolia is 0-indexed but pages are 1-indexed so we need to subtract 1
        params.set("page", (direction - 1).toString());
        return params;
    }

    const newPage = direction === "next" ? currentPage + 1 : Math.max(currentPage - 1, 0);
    params.set("page", newPage.toString());
    return params;
}