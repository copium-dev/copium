import { writable } from "svelte/store";

const getInitialView = () => {
    if (typeof window !== "undefined" && window.localStorage) {
        const savedView = localStorage.getItem("view_preference");
        return savedView === "true" ? true : false;
    }
    return false;
};

export const isGridView = writable(getInitialView());

if (typeof window !== "undefined" && window.localStorage) {
    isGridView.subscribe((value) => {
        localStorage.setItem("view_preference", value.toString());
    });
}
