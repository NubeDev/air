import { } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { DynamicField } from './DynamicField';
import type { JSONSchema } from '@/types/api';

interface SchemaFormProps {
  schema: JSONSchema;
  onSubmit: (data: Record<string, any>) => void;
  initialData?: Record<string, any>;
  isLoading?: boolean;
}

export function SchemaForm({ schema, onSubmit, initialData = {}, isLoading = false }: SchemaFormProps) {
  // Create Zod schema from JSON Schema
  const createZodSchema = (jsonSchema: JSONSchema) => {
    const shape: Record<string, z.ZodTypeAny> = {};
    
    Object.entries(jsonSchema.properties).forEach(([key, prop]) => {
      let zodType: z.ZodTypeAny;
      
      switch (prop.type) {
        case 'string':
          zodType = z.string();
          if (prop.format === 'date') {
            zodType = z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Invalid date format (YYYY-MM-DD)');
          }
          break;
        case 'number':
          zodType = z.number();
          if (prop.minimum !== undefined) {
            zodType = (zodType as z.ZodNumber).min(prop.minimum);
          }
          if (prop.maximum !== undefined) {
            zodType = (zodType as z.ZodNumber).max(prop.maximum);
          }
          break;
        case 'integer':
          zodType = z.number().int();
          if (prop.minimum !== undefined) {
            zodType = (zodType as z.ZodNumber).min(prop.minimum);
          }
          if (prop.maximum !== undefined) {
            zodType = (zodType as z.ZodNumber).max(prop.maximum);
          }
          break;
        case 'boolean':
          zodType = z.boolean();
          break;
        default:
          zodType = z.string();
      }
      
      if (jsonSchema.required.includes(key)) {
        shape[key] = zodType;
      } else {
        shape[key] = zodType.optional();
      }
    });
    
    return z.object(shape);
  };

  const zodSchema = createZodSchema(schema);
  const { control, handleSubmit, formState: { errors } } = useForm({
    resolver: zodResolver(zodSchema),
    defaultValues: initialData,
  });

  const handleFormSubmit = (data: Record<string, any>) => {
    onSubmit(data);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Report Parameters</CardTitle>
        <CardDescription>
          Configure the parameters for your report execution
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-6">
          <div className="grid gap-4">
            {Object.entries(schema.properties).map(([key, prop]) => (
              <DynamicField
                key={key}
                name={key}
                field={prop}
                control={control}
                error={errors[key]?.message}
              />
            ))}
          </div>
          
          <div className="flex justify-end space-x-2">
            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Executing...' : 'Execute Report'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
