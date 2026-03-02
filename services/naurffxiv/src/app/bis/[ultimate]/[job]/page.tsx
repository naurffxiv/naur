import { redirect } from "next/navigation";
import { getBisConfig } from "@/config/bis.config";

interface BisRedirectPageProps {
  params: Promise<{
    ultimate: string;
    job: string;
  }>;
}

/**
 * Dynamic redirect route for BiS links
 * Route: /bis/[ultimate]/[job]
 * Example: /bis/top/pld -> redirects to xivgear.app
 */
export default async function BisRedirectPage({
  params,
}: BisRedirectPageProps): Promise<void> {
  const { ultimate, job } = await params;
  const bisConfig = await getBisConfig();

  // Attempt case-insensitive lookup
  const ultimateConfig = bisConfig[ultimate.toUpperCase()];

  if (!ultimateConfig) {
    redirect("/bis");
  }

  const jobInfo = ultimateConfig.jobs[job.toUpperCase()];

  if (!jobInfo || !jobInfo.xivGearUrl) {
    redirect("/bis");
  }

  // Redirect to the external BiS URL
  redirect(jobInfo.xivGearUrl);
}
