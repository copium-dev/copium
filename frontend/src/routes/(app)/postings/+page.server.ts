import type { PageServerLoad } from './$types';
import { BACKEND_URL } from '$env/static/private';
import { redirect } from '@sveltejs/kit';
import { PUBLIC_LOGO_KEY } from '$env/static/public';
    
interface Posting {
    company_name: string;
    title: string;
    date_updated: number;
    date_posted: number;
    locations: string[];
    url: string;
}

export const load: PageServerLoad = async ({ fetch, url, locals }) => {
    const page = url.searchParams.get('page');
    const query = url.searchParams.get('q');
    const company = url.searchParams.get('company');
    const location = url.searchParams.get('location');
    const title = url.searchParams.get('role');
    const startDate = url.searchParams.get('startDate');
    const endDate = url.searchParams.get('endDate');
    const hitsPerPage = url.searchParams.get('hits');

    const params = new URLSearchParams();
    if (page) params.set('page', page);
    if (query) params.set('q', query);
    if (company) params.set('company', company);
    if (title) params.set('title', title);
    if (location) params.set('location', location);
    if (startDate) params.set('startDate', startDate);
    if (endDate) params.set('endDate', endDate);
    if (hitsPerPage) params.set('hits', hitsPerPage);

    const dashboardURL = `${BACKEND_URL}/postings?${params.toString()}`;

    const response = await fetch(dashboardURL, {
        headers: {
            'Authorization': `Bearer ${locals.authToken}`
        }
    });
    
    if (!response.ok) {
        throw redirect(303, `${BACKEND_URL}/auth/google`);
    }

    const data = await response.json();
    const postings = (data.postings || []) as Posting[];
    const currentPage = parseInt(data.currentPage) || 1;
    const totalPages = parseInt(data.totalPages) || 1;

    // map company names to logo URLs
    const companyNames: string[] = [...new Set(postings.map((posting) => posting.company_name))];
    const logoMap = new Map();
    const logoPromises = companyNames.map(async (company) => {
        try {
            const res = await fetch(
                `https://api.brandfetch.io/v2/search/${encodeURIComponent(company)}?c=${PUBLIC_LOGO_KEY}`
            );
            
            if (res.ok) {
                const data = await res.json();
                const logo = data.length > 0 ? data[0].icon : null;
                logoMap.set(company, logo);
            } else {
                logoMap.set(company, null);
            }
        } catch (error) {
            console.error(`Error fetching logo for ${company}:`, error);
            logoMap.set(company, null);
        }
    });

    await Promise.all(logoPromises);

    const companyLogos = Object.fromEntries(logoMap);
    
    return {
        postings,
        currentPage,
        totalPages,
        companyLogos, 
    };
    
};