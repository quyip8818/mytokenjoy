import {
  ApiCompatibility,
  Capabilities,
  Challenges,
  DeploymentModes,
  DownloadCTA,
  Mission,
  QuotaControl,
  Solutions,
  Testimonials,
} from '@/sections'
import { Footer } from '@/shared'
import type { HomeContent } from '@/content'

export interface HomeRestProps {
  content: HomeContent
}

export default function HomeRest({ content }: HomeRestProps) {
  return (
    <>
      <Challenges content={content.challenges} />
      <Solutions content={content.solutions} />
      <Capabilities content={content.capabilities} />
      <QuotaControl content={content.quota} />
      <ApiCompatibility content={content.integration} />
      <DeploymentModes content={content.deployment} />
      <Testimonials content={content.testimonials} />
      <Mission content={content.mission} />
      <DownloadCTA content={content.cta} />
      <Footer content={content.footer} />
    </>
  )
}
