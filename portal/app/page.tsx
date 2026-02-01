'use client'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useModels } from '@/lib/hooks/useModels'

export default function Home() {
  const { data, isLoading, error } = useModels({ limit: 6 });
  
  const models = data?.data?.models?.map(model => ({
    name: model.name,
    provider: model.provider_name,
    category: model.category || 'Chat',
    price: `$${model.input_price.toFixed(5)} input / $${model.output_price.toFixed(5)} output per token`,
    context: model.context_length ? `${model.context_length.toLocaleString()} tokens` : 'N/A',
    capabilities: Object.keys(model.capabilities || {}),
  })) || [];

  return (
    <>
      {/* Hero Section */}
      <section className="py-20 px-4 text-center">
        <h1 className="text-5xl font-bold tracking-tight mb-6">
          Unified Access to <span className="text-primary">300+</span> AI Models
        </h1>
        <p className="text-xl text-muted-foreground max-w-3xl mx-auto mb-10">
          MassRouter SaaS provides a single API for all major AI models. Streamline your development with unified pricing, billing, and monitoring.
        </p>
        <div className="flex gap-4 justify-center">
          <Button size="lg">Get Started</Button>
          <Button size="lg" variant="outline">View Documentation</Button>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-16 px-4 bg-muted/30">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-3xl font-bold text-center mb-12">Why Choose MassRouter SaaS?</h2>
          <div className="grid md:grid-cols-3 gap-8">
            <Card>
              <CardHeader>
                <CardTitle>Unified API</CardTitle>
                <CardDescription>One interface for all models</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Access 60+ providers with a single API key. No need to manage multiple accounts or integrations.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Smart Routing</CardTitle>
                <CardDescription>Automatic failover and load balancing</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Intelligent routing ensures high availability and optimal performance across all providers.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Usage Analytics</CardTitle>
                <CardDescription>Detailed insights and reporting</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Monitor usage, costs, and performance across all models with comprehensive dashboards.
                </p>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      {/* Models Section */}
      <section className="py-16 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="flex justify-between items-center mb-8">
            <h2 className="text-3xl font-bold">Popular Models</h2>
            <Button variant="outline">View All Models</Button>
          </div>
          {isLoading ? (
            <div className="text-center py-10">
              <p className="text-muted-foreground">Loading models...</p>
            </div>
          ) : error ? (
            <div className="text-center py-10">
              <p className="text-red-500">Failed to load models. Please try again later.</p>
              <p className="text-sm text-muted-foreground mt-2">Error: {error.message}</p>
            </div>
          ) : (
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
              {models.map((model, index) => (
                <Card key={index} className="hover:shadow-lg transition-shadow">
                  <CardHeader>
                    <div className="flex justify-between items-start">
                      <div>
                        <CardTitle>{model.name}</CardTitle>
                        <CardDescription>{model.provider}</CardDescription>
                      </div>
                      <Badge variant="secondary">{model.category}</Badge>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div>
                        <p className="text-sm text-muted-foreground">Price</p>
                        <p className="text-2xl font-bold">{model.price}</p>
                      </div>
                      <div>
                        <p className="text-sm text-muted-foreground">Context Length</p>
                        <p className="font-medium">{model.context}</p>
                      </div>
                      <div>
                        <p className="text-sm text-muted-foreground mb-2">Capabilities</p>
                        <div className="flex flex-wrap gap-2">
                          {model.capabilities.map((cap, i) => (
                            <Badge key={i} variant="outline">{cap}</Badge>
                          ))}
                        </div>
                      </div>
                      <Button className="w-full">Try Now</Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </section>



    </>
  )
}