import type { PageServerLoad } from './$types';
import { BACKEND_URL } from '$env/static/private';
import { redirect } from '@sveltejs/kit';
import { PUBLIC_LOGO_KEY } from '$env/static/public';
import placeholder from '$lib/images/placeholder.png';

export const load: PageServerLoad = async ({ fetch, url }) => {
    const page = url.searchParams.get('page');
    const query = url.searchParams.get('q');
    const company = url.searchParams.get('company');
    const location = url.searchParams.get('location');
    const title = url.searchParams.get('title');
    const startDate = url.searchParams.get('startDate');
    const endDate = url.searchParams.get('endDate');
    const active = url.searchParams.get('active');

    const params = new URLSearchParams();
    if (page) params.set('page', page);
    if (query) params.set('q', query);
    if (company) params.set('company', company);
    if (title) params.set('title', title);
    if (location) params.set('location', location);
    if (startDate) params.set('startDate', startDate);
    if (endDate) params.set('endDate', endDate);
    console.log(params.toString());

    // this should actually be implemented lololol
    const dashboardURL = `${BACKEND_URL}/postings?${params.toString()}`;

    const response = await fetch(dashboardURL, {
        credentials: 'include'  // every protected route needs to include credentials
    });
    
    if (!response.ok) {
        throw redirect(303, `${BACKEND_URL}/auth/google`);
    }
    
    // sample response 
    // const response = {
    //     ok: true,
    //     json: () => {
    //         return {
    //             postings: [
    //                 {
    //                     company_name: "Google",
    //                     title: "Software Engineer",
    //                     date_updated: 1620000000,
    //                     date_posted: 1610000000,
    //                     locations: ["Mountain View, CA"],
    //                     url: "https://careers.google.com/jobs/results/1234567890-software-engineer"
    //                 },
    //                 {
    //                     company_name: "Facebook",
    //                     title: "Product Manager",
    //                     date_updated: 1620000000,
    //                     date_posted: 1610000000,
    //                     locations: ["Menlo Park, CA"],
    //                     url: "https://www.facebook.com/careers/jobs/1234567890-product-manager"
    //                 },
    //                 {
    //                     company_name: "Apple",
    //                     title: "Software Engineer",
    //                     date_updated: 1620000000,
    //                     date_posted: 1610000000,
    //                     locations: ["Cupertino, CA"],
    //                     url: "https://jobs.apple.com/en-us/details/1234567890-software-engineer"
    //                 },
    //                 {
    //                     company_name: "Amazon",
    //                     title: "Product Manager",
    //                     date_updated: 1620000000,
    //                     date_posted: 1610000000,
    //                     locations: ["Seattle, WA"],
    //                     url: "https://www.amazon.jobs/en/jobs/1234567890-product-manager"
    //                 },
    //                 {
    //                     company_name: "Capital One",
    //                     title: "Software Engineer",
    //                     date_updated: 1620000000,
    //                     date_posted: 1610000000,
    //                     locations: ["McLean, VA", "Richmond, VA", "Plano, TX", "New York, NY"],
    //                     url: "https://www.capitalonecareers.com/job/software-engineer-J3W5ZD6JZ3Z2Z2Z2Z2Z"
    //                 }
    //             ],
    //             currentPage: '1',
    //             totalPages: '1'
    //         }
    //     },
    //     currentPage: 1,
    //     totalPages: 1
    // }

    const data = await response.json();
    const postings = (data.postings || []) as Posting[];
    const currentPage = parseInt(data.currentPage) || 1;
    const totalPages = parseInt(data.totalPages) || 1;
    
    // Get unique company names
    interface Posting {
        company_name: string;
        title: string;
        date_updated: number;
        date_posted: number;
        locations: string[];
        url: string;
    }
    
    const companyNames: string[] = [...new Set(postings.map((posting) => posting.company_name))];
    
    // Fetch logos for all companies
    const logoMap = new Map();
    
    // Create promises for all logo fetch requests
    const logoPromises = companyNames.map(async (company) => {
        try {
            const res = await fetch(
                `https://api.brandfetch.io/v2/search/${encodeURIComponent(company)}?c=${PUBLIC_LOGO_KEY}`
            );
            
            if (res.ok) {
                const data = await res.json();
                const logo = data.length > 0 ? data[0].icon : placeholder;
                logoMap.set(company, logo);
            } else {
                logoMap.set(company, placeholder);
            }
        } catch (error) {
            console.error(`Error fetching logo for ${company}:`, error);
            logoMap.set(company, placeholder);
        }
    });
    
    // Wait for all logo fetches to complete
    await Promise.all(logoPromises);
    
    // Convert Map to regular object for serialization
    const companyLogos = Object.fromEntries(logoMap);
    
    return {
        postings,
        currentPage,
        totalPages,
        companyLogos, // Send logo URLs to the client
    };
};