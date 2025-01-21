// import type { PageServerLoad } from './$types';
// import { BACKEND_URL } from '$env/static/private';
// import { redirect } from '@sveltejs/kit';

// export const load: PageServerLoad = async ({ fetch }) => {
//     const response = await fetch(`${BACKEND_URL}/user/profile`, {
//         credentials: 'include'  // every protected route needs to include credentials
//     });
    
//     if (!response.ok) {
//         throw redirect(303, `${BACKEND_URL}/auth/google`);
//     }

//     const data = await response.json();
    
//     return {
//         email: data.email
//     };
// };